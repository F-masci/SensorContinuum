#!/bin/bash

set -e

# Configurazioni
AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}" # "aws" o "localstack"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
COMPOSE_DIR="compose"
COMPOSE_FILE_NAME="mqtt-broker.yml"

# Endpoint per LocalStack
ENDPOINT_URL=""
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy on LocalStack..."
else
  echo "Deploy on AWS..."
fi

# Scarica il file compose
echo "Scarico $COMPOSE_FILE_NAME da s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_FILE_NAME..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_FILE_NAME" "$COMPOSE_FILE_NAME"

# Avvia docker-compose
echo "Avvio docker-compose..."
docker-compose -f "$COMPOSE_FILE_NAME" --env-file "./.env" up -d

echo "docker-compose avviato con successo."
