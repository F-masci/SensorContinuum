#!/bin/bash

# Questo script crea una VPC e un Security Group SSH tramite CloudFormation.
# Puoi scegliere dove fare il deploy (AWS o LocalStack) e personalizzare nome VPC, CIDR della rete e CIDR SSH.
# Esempio d'uso:
#   ./create_region.sh --region eu-west-1 --vpc-name TestVPC --vpc-cidr 10.1.0.0/16 --ssh-cidr 192.168.1.0/24
# Per deploy su LocalStack aggiungi: --deploy=localstack
# Puoi aggiungere altri template modificando le variabili TEMPLATE_xxx.

show_help() {
  echo "Utilizzo: $0 [opzioni]"
  echo "  --deploy=localstack      Deploy su LocalStack invece che AWS"
  echo "  --stack-name NAME        Nome dello stack CloudFormation"
  echo "  --region REGION          Regione AWS (default: eu-east-1)"
  echo "  --vpc-name NAME          Nome della VPC (default: MyVPC)"
  echo "  --vpc-cidr CIDR          CIDR della VPC (default: 10.0.0.0/16)"
  echo "  --ssh-cidr CIDR          CIDR per accesso SSH (default: 0.0.0.0/0)"
  echo "  -h, --help               Mostra questo messaggio"
  echo "Esempio:"
  echo "  $0 --region eu-west-1 --vpc-name TestVPC --vpc-cidr 10.1.0.0/16 --ssh-cidr 192.168.1.0/24"
}

# Definizione dei template da usare
TEMPLATE_VPC="../terraform/region/VPC.yaml"
# Puoi aggiungere altri template qui, ad esempio:
# TEMPLATE_SUBNET="../terraform/region/Subnet.yaml"

AWS_REGION="eu-east-1"
STACK_NAME="${AWS_REGION}_region"
DEPLOY_MODE="aws"
VPC_NAME="${AWS_REGION}_vpc"
VPC_CIDR="10.0.0.0/16"
SSH_CIDR="0.0.0.0/0"

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
    --stack-name)
      STACK_NAME="$2"
      shift 2
      ;;
    --region)
      AWS_REGION="$2"
      VPC_NAME="${AWS_REGION}_vpc"
      STACK_NAME="${AWS_REGION}_region"
      shift 2
      ;;
    --vpc-name)
      VPC_NAME="$2"
      shift 2
      ;;
    --vpc-cidr)
      VPC_CIDR="$2"
      shift 2
      ;;
    --ssh-cidr)
      SSH_CIDR="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

# Imposta endpoint per LocalStack se richiesto
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy su LocalStack..."
else
  ENDPOINT_URL=""
  echo "Deploy su AWS..."
fi

# Deploy del template VPC
echo "Deploy del template VPC..."
echo "Parametri: VPC Name=$VPC_NAME, VPC CIDR=$VPC_CIDR, SSH CIDR=$SSH_CIDR"
aws cloudformation deploy \
  --template-file "$TEMPLATE_VPC" \
  --stack-name "$STACK_NAME" \
  --region "$AWS_REGION" \
  --capabilities CAPABILITY_NAMED_IAM \
  --parameter-overrides VpcName="$VPC_NAME" VpcCidr="$VPC_CIDR" SshAllowedCidr="$SSH_CIDR" \
  $ENDPOINT_URL
if [[ $? -ne 0 ]]; then
  echo "Errore nel deploy del template VPC."
  exit 1
fi

aws ec2 $ENDPOINT_URL describe-vpcs --region $AWS_REGION --filters "Name=tag:Name,Values=$VPC_NAME" --query "Vpcs[0].VpcId" --output text

# Esempio di deploy di un altro template (decommenta e personalizza se necessario)
# echo "Deploy del template Subnet..."
# aws cloudformation deploy \
#   --template-file "$TEMPLATE_SUBNET" \
#   --stack-name "${STACK_NAME}_subnet" \
#   --region "$AWS_REGION" \
#   --capabilities CAPABILITY_NAMED_IAM \
#   --parameter-overrides ... \
#   $ENDPOINT_URL

echo "Deploy completato per tutti i template."