#!/bin/bash

set -e

# Configurazioni
AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}" # "aws" o "localstack"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
COMPOSE_DIR="compose"
ENV_SUBDIR="$COMPOSE_DIR/envs/region-001/build-0001"
ENV_FILE_NAME=".env.floor001"
COMPOSE_FILE_NAME="edge-hub.yaml"

# Endpoint per LocalStack
ENDPOINT_URL=""
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy su LocalStack..."
else
  echo "Deploy su AWS..."
fi

# Crea cartelle locali se non esistono
mkdir -p "$COMPOSE_DIR/$ENV_SUBDIR"

# Scarica il file compose
echo "Scarico $COMPOSE_FILE_NAME da s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_FILE_NAME..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_FILE_NAME" "$COMPOSE_DIR/$COMPOSE_FILE_NAME"

# Scarica il file .env
echo "Scarico $ENV_FILE_NAME da s3://$BUCKET_NAME/$ENV_SUBDIR/$ENV_FILE_NAME..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$ENV_SUBDIR/$ENV_FILE_NAME" "$COMPOSE_DIR/$ENV_SUBDIR/$ENV_FILE_NAME"

# Avvia docker-compose
echo "Avvio docker-compose..."
docker-compose -f "$COMPOSE_DIR/$COMPOSE_FILE_NAME" --env-file "$COMPOSE_DIR/$ENV_SUBDIR/$ENV_FILE_NAME" up -d

echo "docker-compose avviato con successo."
