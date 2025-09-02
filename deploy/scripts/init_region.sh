#!/bin/bash

TEMPLATE_FILE=${1:-../terraform/region.yaml}
AWS_REGION=${2:-eu-east-1}
STACK_NAME="${AWS_REGION}_region"
DEPLOY_MODE="aws"

for arg in "$@"; do
  if [[ "$arg" == "--deploy=localstack" ]]; then
    DEPLOY_MODE="localstack"
  elif [[ "$arg" == "--stack-name" ]]; then
    STACK_NAME="${!i+1}"
  elif [[ "$arg" == "--region" ]]; then
    REGION="${!i+1}"
  fi
done

if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy su LocalStack..."
else
  ENDPOINT_URL=""
  echo "Deploy su AWS..."
fi

aws cloudformation deploy \
  --template-file "$TEMPLATE_FILE" \
  --stack-name "$STACK_NAME" \
  --region "$AWS_REGION" \
  --capabilities CAPABILITY_NAMED_IAM \
  $ENDPOINT_URL

echo "Stack creato dal template $TEMPLATE_FILE nella regione $AWS_REGION ($DEPLOY_MODE)."