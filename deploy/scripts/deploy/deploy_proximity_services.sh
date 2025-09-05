#!/bin/bash

set -e

# Configurazioni
AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}" # "aws" o "localstack"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
COMPOSE_DIR="compose"
COMPOSE_PROXIMITY_HUB_FILE_NAME="proximity-fog-hub.yaml"

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
echo "Scarico $COMPOSE_PROXIMITY_HUB_FILE_NAME da s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_PROXIMITY_HUB_FILE_NAME..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_PROXIMITY_HUB_FILE_NAME" "$COMPOSE_PROXIMITY_HUB_FILE_NAME"
if [ $? -ne 0 ]; then
  echo "Errore nel download di $COMPOSE_PROXIMITY_HUB_FILE_NAME"
  exit 1
fi

echo "Controllo se esiste il volume macrozone-hub-${EDGE_MACROZONE}-cache-data..."
if [ ! "$(docker volume ls -q -f name=macrozone-hub-${EDGE_MACROZONE}-cache-data)" ]; then
  echo "Volume macrozone-hub-${EDGE_MACROZONE}-cache-data non trovato. Creo il volume..."
  docker volume create macrozone-hub-${EDGE_MACROZONE}-cache-data || { echo "Errore nella creazione del volume"; exit 1; }
else
  echo "Volume macrozone-hub-${EDGE_MACROZONE}-cache-data già esistente."
fi

# Avvia proximity-fog-hub
if [ ! -f "$COMPOSE_PROXIMITY_HUB_FILE_NAME" ]; then
  echo "File $COMPOSE_PROXIMITY_HUB_FILE_NAME non trovato, esco."
  exit 1
else
  echo "Avvio proximity-fog-hub..."
  docker-compose -f "$COMPOSE_PROXIMITY_HUB_FILE_NAME" --env-file ".env" -p proximity-fog-hub up -d
fi

# Aumenta il numero massimo di connessioni di postgres
echo "Aumento il numero massimo di connessioni di postgres a 500..."
docker exec -it macrozone-hub-${EDGE_MACROZONE}-cache psql -U admin -d sensorcontinuum -c "ALTER SYSTEM SET max_connections = 500;"
# Riavvia il progetto docker-compose per applicare la modifica
docker-compose -p proximity-fog-hub restart

echo "docker-compose avviato con successo."
