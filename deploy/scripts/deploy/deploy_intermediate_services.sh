#!/bin/bash

set -e

# Configurazioni
AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}" # "aws" o "localstack"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
COMPOSE_DIR="compose"
COMPOSE_INTERMEDIATE_HUB_FILE_NAME="intermediate-fog-hub.yaml"

# Endpoint per LocalStack
ENDPOINT_URL=""
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy on LocalStack..."
else
  echo "Deploy on AWS..."
fi

# Carica variabili d'ambiente
if [ -f "./.env" ]; then
  source ./.env
  echo "Variabili d'ambiente caricate da .env"
else
  echo "File .env non trovato!"
  exit 1
fi

# Scarica il file compose
echo "Scarico $COMPOSE_INTERMEDIATE_HUB_FILE_NAME da s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_INTERMEDIATE_HUB_FILE_NAME..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_INTERMEDIATE_HUB_FILE_NAME" "$COMPOSE_INTERMEDIATE_HUB_FILE_NAME"
if [ $? -ne 0 ]; then
  echo "Errore nel download di $COMPOSE_INTERMEDIATE_HUB_FILE_NAME"
  exit 1
fi

# Avvia intermediate-fog-hub
if [ ! -f "$COMPOSE_INTERMEDIATE_HUB_FILE_NAME" ]; then
  echo "File $COMPOSE_INTERMEDIATE_HUB_FILE_NAME non trovato, esco."
  exit 1
else
  echo "Avvio intermediate-fog-hub..."
  docker-compose -f "$COMPOSE_INTERMEDIATE_HUB_FILE_NAME" --env-file ".env" -p intermediate-fog-hub up -d
fi

echo "docker-compose avviato con successo."
