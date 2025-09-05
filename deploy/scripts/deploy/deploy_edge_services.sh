#!/bin/bash

set -e

# Configurazioni
AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}" # "aws" o "localstack"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
COMPOSE_DIR="compose"
COMPOSE_EDGE_HUB_FILE_NAME="edge-hub.yaml"
COMPOSE_SENSORS_FILE_NAME="sensor-agent.generated_20_1.yml"

# Endpoint per LocalStack
ENDPOINT_URL=""
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy on LocalStack..."
else
  echo "Deploy on AWS..."
fi

# Carica variabili d'ambiente
if [ -f ".env" ]; then
  source .env
  echo "Variabili d'ambiente caricate da .env"
else
  echo "File .env non trovato!"
  exit 1
fi

# Scarica il file compose
echo "Scarico $COMPOSE_EDGE_HUB_FILE_NAME da s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_EDGE_HUB_FILE_NAME..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_EDGE_HUB_FILE_NAME" "$COMPOSE_EDGE_HUB_FILE_NAME"
if [ $? -ne 0 ]; then
  echo "Errore nel download di $COMPOSE_EDGE_HUB_FILE_NAME"
  exit 1
fi

echo "Scarico $COMPOSE_SENSORS_FILE_NAME da s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_SENSORS_FILE_NAME..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_SENSORS_FILE_NAME" "$COMPOSE_SENSORS_FILE_NAME"
if [ $? -ne 0 ]; then
  echo "Errore nel download di $COMPOSE_SENSORS_FILE_NAME"
  exit 1
fi

# Avvia edge-hub
if [ -f "$COMPOSE_EDGE_HUB_FILE_NAME" ]; then
  echo "Avvio edge-hub..."
  docker-compose -f "$COMPOSE_EDGE_HUB_FILE_NAME" --env-file ".env" -p edge-hub up -d
else
  echo "File $COMPOSE_EDGE_HUB_FILE_NAME non trovato, esco."
  exit 1
fi

# Avvio sensori
if [ -f "$COMPOSE_SENSORS_FILE_NAME" ]; then
  echo "Avvio sensori..."
  docker-compose -f "$COMPOSE_SENSORS_FILE_NAME" --env-file ".env" -p sensors up -d
else
  echo "File $COMPOSE_SENSORS_FILE_NAME non trovato, esco."
  exit 1
fi

# Riga cron che vuoi aggiungere
CRON_JOB="0 3 * * * /usr/bin/docker docker-compose -p sensors restart >> /var/log/sensors-restart.log 2>&1"

# Controlla se esiste già, se no lo aggiunge
( crontab -l 2>/dev/null | grep -F "$CRON_JOB" ) || \
( crontab -l 2>/dev/null; echo "$CRON_JOB" ) | crontab -

echo "Zona avviata con successo."
