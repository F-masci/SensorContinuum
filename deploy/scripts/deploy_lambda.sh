#!/bin/bash
set -e

STACK_NAME=$1       # nome dello stack CloudFormation
FOLDER=$2
FUNCTION=$3
PATH_ROUTE=$4       # percorso della route, es: /zone/sensor/data/raw/{region}/{macrozone}/{zone}/{sensor}
RESET=$5            # opzionale: --reset

if [ -z "$FOLDER" ] || [ -z "$FUNCTION" ] || [ -z "$PATH_ROUTE" ] || [ -z "$STACK_NAME" ]; then
  echo "Usage: $0 <stack_name> <folder> <function> <path_route> [--reset]"
  exit 1
fi

CLOUDFORMATION_DIR="../cloudformation"
BASE_TEMPLATE="$CLOUDFORMATION_DIR/lambda.template.yaml"
OUTPUT_DIR="$CLOUDFORMATION_DIR/lambda/$FOLDER"
OUTPUT_FILE="$OUTPUT_DIR/$FUNCTION.yaml"

mkdir -p "$OUTPUT_DIR"

# --- Generazione template ---
SECURITY_GROUP_ID=$(aws ec2 describe-security-groups --filters Name=tag:Name,Values=sc-sg-lambda --query 'SecurityGroups[0].GroupId' --output text)
SUBNET_ID_1=$(aws ec2 describe-subnets --filters Name=tag:Name,Values=sc-subnet-lambda-private-1 --query 'Subnets[0].SubnetId' --output text)
SUBNET_ID_2=$(aws ec2 describe-subnets --filters Name=tag:Name,Values=sc-subnet-lambda-private-2 --query 'Subnets[0].SubnetId' --output text)

echo "[INFO] Creo template per $FUNCTION"
sed \
  -e "s|\${FUNCTION_NAME}|$FUNCTION|g" \
  -e "s|\${FOLDER_PATH}|$FOLDER|g" \
  -e "s|\${SECURITY_GROUP_ID}|$SECURITY_GROUP_ID|g" \
  -e "s|\${SUBNET_ID_1}|$SUBNET_ID_1|g" \
  -e "s|\${SUBNET_ID_2}|$SUBNET_ID_2|g" \
  "$BASE_TEMPLATE" > "$OUTPUT_FILE"

# --- Build ---
echo "[INFO] Eseguo sam build"
sam build --template-file "$OUTPUT_FILE"

# --- Controllo e creazione bucket S3 ---
BUCKET_NAME="sensor-continuum-lambda"
if ! aws s3api head-bucket --bucket "$BUCKET_NAME" 2>/dev/null; then
  echo "[INFO] Il bucket $BUCKET_NAME non esiste. Lo creo..."
  aws s3api create-bucket --bucket "$BUCKET_NAME"
else
  echo "[INFO] Il bucket $BUCKET_NAME esiste già."
fi

# --- Deploy ---
echo "[INFO] Avvio deploy con SAM"
sam deploy \
  --template-file "$OUTPUT_FILE" \
  --stack-name "$STACK_NAME" \
  --s3-bucket "$BUCKET_NAME"

# --- Recupera API ID ---
API_NAME="Sensor Continuum API"
echo "[INFO] Recupero API ID"
API_ID=$(aws apigatewayv2 get-apis --query "Items[?Name=='$API_NAME'].ApiId" --output text)
if [ -z "$API_ID" ]; then
  echo "[ERROR] API ID non trovato. Assicurati che l'API '$API_NAME' esista."
  exit 1
fi
echo "[INFO] API ID trovato: $API_ID"

# --- Aggiungi permessi Lambda per API Gateway ---
echo "[INFO] Controllo se i permessi per API Gateway esistono già"
EXISTING_PERMISSION=$(aws lambda get-policy --function-name "$FUNCTION" --query 'Policy' --output text | grep apigateway.amazonaws.com || true)
if [ -n "$EXISTING_PERMISSION" ] && [ "$RESET" != "--reset" ]; then
  echo "[INFO] I permessi per API Gateway esistono già. Usa --reset per rigenerarli."
else
  if [ "$RESET" == "--reset" ]; then
    echo "[INFO] Rimuovo permessi esistenti per API Gateway"
    STATEMENT_IDS=$(aws lambda get-policy --function-name "$FUNCTION" --query 'Policy' --output text | grep apigateway.amazonaws.com | jq -r '.Statement[].Sid')
    for SID in $STATEMENT_IDS; do
      aws lambda remove-permission --function-name "$FUNCTION" --statement-id "$SID" || true
    done
  fi
  echo "[INFO] Aggiungo permessi Lambda per API Gateway"
  aws lambda add-permission \
    --function-name "$FUNCTION" \
    --statement-id "apigw-invoke-$(date +%s)" \
    --action lambda:InvokeFunction \
    --principal apigateway.amazonaws.com \
    --source-arn "arn:aws:execute-api:us-east-1:975050105348:$API_ID/*/GET$PATH_ROUTE"
fi

# --- Crea integrazione Lambda HTTP API ---
echo "[INFO] Controllo se la route GET $PATH_ROUTE esiste già"
EXISTING_ROUTE=$(aws apigatewayv2 get-routes --api-id "$API_ID" --query "Items[?RouteKey=='GET $PATH_ROUTE'].RouteId" --output text || true)

echo "[INFO] Controllo se l'integrazione esiste già"
EXISTING_INTEGRATION=$(aws apigatewayv2 get-integrations --api-id "$API_ID" --query "Items[?IntegrationUri=='arn:aws:lambda:us-east-1:975050105348:function:$FUNCTION'].IntegrationId" --output text || true)

if [ -n "$EXISTING_INTEGRATION" ] && [ "$RESET" != "--reset" ]; then
  echo "[INFO] L'integrazione esiste già con ID: $EXISTING_INTEGRATION. Usa --reset per rigenerarla."
  INTEGRATION_ID=$EXISTING_INTEGRATION
else
  if [ "$RESET" == "--reset" ]; then
    echo "[INFO] Rimuovo route esistente"
    if [ -n "$EXISTING_ROUTE" ]; then
      aws apigatewayv2 delete-route --api-id "$API_ID" --route-id "$EXISTING_ROUTE" || true
    fi
    echo "[INFO] Rimuovo integrazione esistente"
    if [ -n "$EXISTING_INTEGRATION" ]; then
      aws apigatewayv2 delete-integration --api-id "$API_ID" --integration-id "$EXISTING_INTEGRATION" || true
    fi
  fi
  echo "[INFO] Creo integrazione HTTP API"
  INTEGRATION_ID=$(aws apigatewayv2 create-integration \
    --api-id "$API_ID" \
    --integration-type AWS_PROXY \
    --integration-uri "arn:aws:lambda:us-east-1:975050105348:function:$FUNCTION" \
    --payload-format-version 2.0 \
    --request-parameters overwrite:path="\$request.path" \
    --query 'IntegrationId' --output text)
  echo "[INFO] Integrazione creata con ID: $INTEGRATION_ID"
fi

# --- Crea route GET sulla API con integrazione ---
if [ -n "$EXISTING_ROUTE" ] && [ "$RESET" != "--reset" ]; then
  echo "[WARNING] La route GET $PATH_ROUTE esiste già con ID: $EXISTING_ROUTE. Usa --reset per rigenerarla."
else
  if [ "$RESET" == "--reset" ]; then
    echo "[INFO] Rimuovo route esistente"
    if [ -n "$EXISTING_ROUTE" ]; then
      aws apigatewayv2 delete-route --api-id "$API_ID" --route-id "$EXISTING_ROUTE" || true
    fi
  fi
  echo "[INFO] Creo route GET $PATH_ROUTE"
  aws apigatewayv2 create-route \
    --api-id "$API_ID" \
    --route-key "GET $PATH_ROUTE" \
    --target "integrations/$INTEGRATION_ID"
fi

echo "[INFO] Lambda $FUNCTION agganciata a HTTP API $API_ID su route GET $PATH_ROUTE con PATH mapping \$request.path"
