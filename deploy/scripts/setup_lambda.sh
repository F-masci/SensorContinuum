#!/bin/bash

# ----------------
# Network setup
# ----------------

STACK_NAME="sc-lambda-network"                          # Nome dello stack CloudFormation
TEMPLATE_FILE="../cloudformation/lambda-network.yaml"   # Percorso al template yaml

echo "[INFO] Deploy stack $STACK_NAME con template $TEMPLATE_FILE"
aws cloudformation deploy \
  --stack-name "$STACK_NAME" \
  --template-file "$TEMPLATE_FILE" \
  --capabilities CAPABILITY_NAMED_IAM

# Recupera e mostra i valori utili per script lambda
VPC_ID=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='VPCId'].OutputValue" --output text)
SG_ID=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='LambdaSecurityGroupId'].OutputValue" --output text)
SUBNET1=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='PrivateSubnet1Id'].OutputValue" --output text)
SUBNET2=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='PrivateSubnet2Id'].OutputValue" --output text)

echo "[INFO] Recupero output dello stack:"
echo " VPC: $VPC_ID"
echo " SecurityGroup: $SG_ID"
echo " Subnet1: $SUBNET1"
echo " Subnet2: $SUBNET2"
echo

#!/bin/bash
set -e

STACK_NAME="sc-lambda-api"
TEMPLATE_FILE="../cloudformation/lambda-api.yaml"
PUBLIC_HOSTED_ZONE_NAME="sensor-continuum.it"

# Recupera Hosted Zone ID
HOSTED_ZONE_ID=$(aws route53 list-hosted-zones-by-name \
  --dns-name "$PUBLIC_HOSTED_ZONE_NAME" \
  --query "HostedZones[0].Id" --output text)
HOSTED_ZONE_ID=${HOSTED_ZONE_ID#/hostedzone/}

# -----------------------------
# Processo background: monitor CloudFormation
# -----------------------------
(

  echo "[BG] Attendo che lo stack entri in CREATE_IN_PROGRESS..."

  sleep 30

  # Loop finché lo stack non è almeno CREATE_IN_PROGRESS
  while true; do
      STATUS=$(aws cloudformation describe-stacks \
          --stack-name "$STACK_NAME" \
          --query "Stacks[0].StackStatus" \
          --output text)

      echo "[BG] Stato attuale dello stack: $STATUS."

      if [[ "$STATUS" == CREATE_COMPLETE* || "$STATUS" == UPDATE_COMPLETE* ]]; then\
        echo "[BG] Stack completato con stato $STATUS, esco dal monitor."
        exit 0
      fi

      if [[ "$STATUS" == CREATE_IN_PROGRESS* ]]; then
          break
      fi

      sleep 2
  done

  echo "[BG] Avvio monitor eventi stack per validazione ACM..."

  while true; do

    echo "[BG] Controllo eventi stack..."

    # Prendi ultimi eventi della stack
    EVENTS=$(aws cloudformation describe-stack-events \
      --stack-name "$STACK_NAME" \
      --query "StackEvents[?ResourceType=='AWS::CertificateManager::Certificate']" \
      --output json)

    # Legge ogni riga JSON in un array
        mapfile -t rows < <(echo "$EVENTS" | jq -c '.[]')

    for row in "${rows[@]}"; do
      STATUS=$(echo "$row" | jq -r '.ResourceStatus')
      REASON=$(echo "$row" | jq -r '.ResourceStatusReason // empty')

      if [[ "$STATUS" == "CREATE_IN_PROGRESS" && "$REASON" == *"Content of DNS Record is:"* ]]; then

        echo "[BG] Trovato evento di CREATE_IN_PROGRESS con Reason: $REASON"

        # Estrae la parte JSON del messaggio
        REASON_JSON=$(echo "$REASON" | sed -n 's/.*Content of DNS Record is: \(.*\)/\1/p')

        # Trasforma in JSON valido aggiungendo virgolette alle chiavi e ai valori
        REASON_JSON=$(echo "$REASON_JSON" | sed 's/\([A-Za-z0-9_]*\):/\"\1\":/g; s/: \([A-Za-z0-9_.-]*\)/: \"\1\"/g; s/,$//')

        # Estrai valori
        NAME=$(echo "$REASON_JSON" | jq -r '.Name')
        TYPE=$(echo "$REASON_JSON" | jq -r '.Type')
        VALUE=$(echo "$REASON_JSON" | jq -r '.Value')

        echo "[BG] Valori estratti: Name=$NAME, Type=$TYPE, Value=$VALUE"

        # Controllo valori
        if [[ -z "$NAME" || -z "$TYPE" || -z "$VALUE" || "$NAME" == "null" || "$TYPE" == "null" || "$VALUE" == "null" ]]; then
            echo "[BG] Errore: impossibile estrarre valori validi da Reason, esco."
            exit 1
        fi

        # Inserisce il record DNS
        echo "[BG] Inserisco record DNS: $NAME $TYPE $VALUE"
        aws route53 change-resource-record-sets \
          --hosted-zone-id "$HOSTED_ZONE_ID" \
          --change-batch "{
            \"Changes\": [{
              \"Action\": \"UPSERT\",
              \"ResourceRecordSet\": {
                \"Name\": \"$NAME\",
                \"Type\": \"$TYPE\",
                \"TTL\": 300,
                \"ResourceRecords\": [{\"Value\": \"$VALUE\"}]
              }
            }]
          }"
        echo "[BG] Record DNS inserito"
        # Una volta fatto termina il processo background
        echo "[BG] Record DNS creato, esco dal monitor."
        exit 0
      fi
    done

    sleep 5

  done
) &

# -----------------------------
# Deploy stack CloudFormation
# -----------------------------
echo "[INFO] Deploy stack $STACK_NAME"
aws cloudformation deploy \
  --stack-name "$STACK_NAME" \
  --template-file "$TEMPLATE_FILE"

# -----------------------------
# Recupera output
# -----------------------------
API_ID=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" \
  --query "Stacks[0].Outputs[?OutputKey=='ApiId'].OutputValue" --output text)

echo "[INFO] API ID: $API_ID"
# -----------------------------
# Recupera il dominio personalizzato API
# -----------------------------

DOMAIN_NAME=$(aws apigatewayv2 get-domain-names \
  --query "Items[?DomainName=='api.sensor-continuum.it'].DomainName" \
  --output text)

# Prendi il record DNS da Route53 di validazione ACM (DomainNameConfigurations)
CNAME_NAME=$(aws apigatewayv2 get-domain-names \
  --query "Items[?DomainName=='$DOMAIN_NAME'].DomainNameConfigurations[0].ApiGatewayDomainName" \
  --output text)

echo "[INFO] Dominio personalizzato API: $DOMAIN_NAME"
echo "[INFO] Record CNAME da aggiungere: $CNAME_NAME"

# -----------------------------
# Aggiungi record CNAME in Route53
# -----------------------------
aws route53 change-resource-record-sets \
  --hosted-zone-id "$HOSTED_ZONE_ID" \
  --change-batch "{
    \"Changes\": [{
      \"Action\": \"UPSERT\",
      \"ResourceRecordSet\": {
        \"Name\": \"$DOMAIN_NAME\",
        \"Type\": \"CNAME\",
        \"TTL\": 300,
        \"ResourceRecords\": [{\"Value\": \"$CNAME_NAME\"}]
      }
    }]
  }"

echo "[INFO] Record CNAME inserito nella zona pubblica."
echo "[INFO] Deploy e configurazione dominio completati."
