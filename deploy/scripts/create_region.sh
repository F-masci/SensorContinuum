#!/bin/bash

# Importa le funzioni
source utils.sh

show_help() {
  echo "Utilizzo: $0 region-name [opzioni]"
  echo "  --deploy=localstack      Deploy su LocalStack invece che AWS"
  echo "  --aws-region REGION      Regione AWS (default: us-east-1)"
  echo "  -h, --help               Mostra questo messaggio"
  echo "Esempio:"
  echo "  $0 region-001 --aws-region eu-east-1"
}

# Definizione dei template da usare
TEMPLATE_VPC="../terraform/region/VPC.yaml"

REGION="$1"
if [[ -z "$REGION" ]]; then
  echo "Errore: il nome della regione è obbligatorio."
  show_help
  exit 1
fi
shift

AWS_REGION="us-east-1"
DEPLOY_MODE="aws"

# Parsing degli argomenti
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      show_help
      exit 0
      ;;
    --deploy=localstack)
      DEPLOY_MODE="localstack"
      shift
      ;;
    --aws-region)
      AWS_REGION="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

STACK_NAME="$REGION-stack"
VPC_NAME="$REGION-vpc"
SUBNET_NAME="$REGION-subnet"
VPC_CIDR="10.0.0.0/16"
SUBNET_CIDR="10.0.0.0/24"
SSH_CIDR="0.0.0.0/0"
HOSTED_ZONE_NAME="$REGION.sensor-continuum.local"

# Imposta endpoint per LocalStack se richiesto
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy su LocalStack..."
else
  ENDPOINT_URL=""
  echo "Deploy su AWS..."
fi

# Deploy del template
echo "Deploy del template VPC..."

echo "Parametri usati:"
echo "  Regione AWS: $AWS_REGION"
echo "  Regione: $REGION"
echo "  Nome Stack: $STACK_NAME"
echo "  Nome VPC: $VPC_NAME"
echo "  Nome Subnet: $SUBNET_NAME"
echo "  CIDR VPC: $VPC_CIDR"
echo "  CIDR Subnet: $SUBNET_CIDR"
echo "  CIDR SSH: $SSH_CIDR"
echo "  Hosted Zone Name: $HOSTED_ZONE_NAME"

aws cloudformation deploy \
  --template-file "$TEMPLATE_VPC" \
  --stack-name "$STACK_NAME" \
  --region "$AWS_REGION" \
  --capabilities CAPABILITY_NAMED_IAM \
  --parameter-overrides VpcName="$VPC_NAME" VpcCidr="$VPC_CIDR" SubnetName="$SUBNET_NAME" SubnetCidr="$SUBNET_CIDR" SshCidr="$SSH_CIDR" HostedZoneName="$HOSTED_ZONE_NAME" \
  $ENDPOINT_URL
if [[ $? -ne 0 ]]; then
  echo "Errore nel deploy del template VPC."
  aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$STACK_NAME"
  exit 1
fi
echo "Deploy del template VPC completato con successo."

# Verifica lo stato dello stack
echo "Verifica lo stato dello stack..."
STACK_STATUS=$(aws cloudformation $ENDPOINT_URL describe-stacks --region "$AWS_REGION" --stack-name "$STACK_NAME" --query "Stacks[0].StackStatus" --output text)
if [[ "$STACK_STATUS" != "CREATE_COMPLETE" && "$STACK_STATUS" != "UPDATE_COMPLETE" ]]; then
  echo "Errore: lo stack $STACK_NAME non è in stato CREATE_COMPLETE o UPDATE_COMPLETE. Stato attuale: $STACK_STATUS"
  aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$STACK_NAME"
  exit 1
fi
echo "Lo stack $STACK_NAME è in stato $STACK_STATUS."

# Verifica la creazione delle risorse
echo "Verifica la creazione delle risorse..."

echo "Verifica della VPC..."
VPC_ID=$(aws ec2 $ENDPOINT_URL describe-vpcs --region "$AWS_REGION" --filters "Name=tag:Name,Values=$VPC_NAME" --query "Vpcs[0].VpcId" --output text)
if [[ -z "$VPC_ID" || "$VPC_ID" == "None" ]]; then
  echo "Errore: VPC con tag Name=$VPC_NAME non trovata."
  exit 1
fi
echo "Trovato VPC ID: $VPC_ID per VPC $VPC_NAME"

echo "Verifica della Subnet..."
SUBNET_ID=$(aws ec2 $ENDPOINT_URL describe-subnets --region "$AWS_REGION" --filters "Name=tag:Name,Values=$SUBNET_NAME" --query "Subnets[0].SubnetId" --output text)
if [[ -z "$SUBNET_ID" || "$SUBNET_ID" == "None" ]]; then
  echo "Errore: subnet con nome $SUBNET_NAME non trovata."
  exit 1
fi
echo "Trovata Subnet ID: $SUBNET_ID per Subnet $SUBNET_NAME"

echo "Verifica della Hosted Zone..."
HOSTED_ZONE_ID=$(aws route53 $ENDPOINT_URL list-hosted-zones-by-name --dns-name "$HOSTED_ZONE_NAME" --query "HostedZones[0].Id" --output text | sed 's|/hostedzone/||')
if [[ -z "$HOSTED_ZONE_ID" || "$HOSTED_ZONE_ID" == "None" ]]; then
  echo "Errore: Hosted Zone con nome $HOSTED_ZONE_NAME non trovata."
  exit 1
fi
echo "Trovata Hosted Zone ID: $HOSTED_ZONE_ID per Hosted Zone $HOSTED_ZONE_NAME"

echo "Deploy completato."