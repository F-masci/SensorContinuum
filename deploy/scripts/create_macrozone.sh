#!/bin/bash

# Script per creare Subnet e Security Group bloccato, con controllo VPC e subnet.
# Esempio:
#   ./create_macrozone.sh --region-name region001 --vpc-name MyVPC --subnet-cidr 10.0.1.0/24 --subnet-name MacrozoneSubnet

show_help() {
  echo "Utilizzo: $0 [opzioni]"
  echo "  --deploy=localstack      Deploy su LocalStack invece che AWS"
  echo "  --stack-name NAME        Nome dello stack CloudFormation"
  echo "  --region-name NAME       Nome logico della regione"
  echo "  --vpc-name NAME          Nome della VPC (tag Name)"
  echo "  --aws-region REGION      Regione AWS (default: eu-west-1)"
  echo "  --subnet-cidr CIDR       CIDR della subnet (default: 10.0.1.0/24)"
  echo "  --subnet-name NAME       Nome della subnet (default: MacrozoneSubnet)"
  echo "  -h, --help               Mostra questo messaggio"
}

SUBNET_TEMPLATE="../terraform/macrozone/Subnet.yaml"
STACK_NAME="${AWS_REGION}_macrozone"
DEPLOY_MODE="aws"
REGION_NAME=""
VPC_NAME=""
AWS_REGION="eu-east-1"
SUBNET_CIDR="10.0.1.0/24"
SUBNET_NAME="${AWS_REGION}_subnet"

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
    --stack-name)
      STACK_NAME="$2"
      shift 2
      ;;
    --region-name)
      REGION_NAME="$2"
      shift 2
      ;;
    --vpc-name)
      VPC_NAME="$2"
      shift 2
      ;;
    --aws-region)
      AWS_REGION="$2"
      SUBNET_NAME="${AWS_REGION}_vpc"
      STACK_NAME="${AWS_REGION}_macrozone"
      shift 2
      ;;
    --subnet-cidr)
      SUBNET_CIDR="$2"
      shift 2
      ;;
    --subnet-name)
      SUBNET_NAME="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

if [[ -z "$REGION_NAME" || -z "$VPC_NAME" ]]; then
  echo "Errore: devi specificare --region-name e --vpc-name"
  exit 1
fi

if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy su LocalStack..."
else
  ENDPOINT_URL=""
  echo "Deploy su AWS..."
fi

# Cerca il VPC ID tramite il tag Name
VPC_ID=$(aws ec2 $ENDPOINT_URL describe-vpcs --region $AWS_REGION --filters "Name=tag:Name,Values=$VPC_NAME" --query "Vpcs[0].VpcId" --output text)
if [[ -z "$VPC_ID" || "$VPC_ID" == "None" ]]; then
  echo "Errore: VPC con tag Name=$VPC_NAME non trovato."
  exit 1
fi
echo "Trovato VPC ID: $VPC_ID per VPC $VPC_NAME"

# Deploy del template Subnet e Security Group
aws cloudformation deploy \
  --template-file "$SUBNET_TEMPLATE" \
  --stack-name "$STACK_NAME" \
  --region "$AWS_REGION" \
  --capabilities CAPABILITY_NAMED_IAM \
  --parameter-overrides VpcId="$VPC_ID" SubnetCidr="$SUBNET_CIDR" SubnetName="$SUBNET_NAME" \
  $ENDPOINT_URL

if [[ $? -ne 0 ]]; then
  echo "Errore nel deploy del template SubnetAndSG."
  exit 1
fi

# Recupera l'ID della subnet creata
SUBNET_ID=$(aws ec2 $ENDPOINT_URL describe-subnets --region "$AWS_REGION" --filters "Name=tag:Name,Values=$SUBNET_NAME" --query "Subnets[0].SubnetId" --output text)
if [[ -z "$SUBNET_ID" || "$SUBNET_ID" == "None" ]]; then
  echo "Errore: subnet $SUBNET_NAME non trovata."
  exit 1
fi

# Controlla che la subnet sia nella VPC corretta
SUBNET_VPC_ID=$(aws ec2 $ENDPOINT_URL describe-subnets --region "$AWS_REGION" --subnet-ids "$SUBNET_ID" --query "Subnets[0].VpcId" --output text)
if [[ "$SUBNET_VPC_ID" != "$VPC_ID" ]]; then
  echo "Errore: la subnet $SUBNET_NAME ($SUBNET_ID) non appartiene alla VPC $VPC_NAME ($VPC_ID)."
  exit 1
fi

echo "Subnet $SUBNET_NAME ($SUBNET_ID) valida nella VPC $VPC_NAME ($VPC_ID)."
echo "Deploy completato."