#!/bin/bash

set -e

# Configurazioni
AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}" # "aws" o "localstack"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
COMPOSE_DIR="compose"
COMPOSE_MQTT_BROKER_FILE_NAME="mqtt-broker.yml"

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
echo "Scarico $COMPOSE_MQTT_BROKER_FILE_NAME da s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_MQTT_BROKER_FILE_NAME..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_MQTT_BROKER_FILE_NAME" "$COMPOSE_MQTT_BROKER_FILE_NAME"
if [ $? -ne 0 ]; then
  echo "Errore nel download di $COMPOSE_MQTT_BROKER_FILE_NAME"
  exit 1
fi

# Avvia mqtt-broker
if [ ! -f "$COMPOSE_MQTT_BROKER_FILE_NAME" ]; then
  echo "File $COMPOSE_MQTT_BROKER_FILE_NAME non trovato, esco."
  exit 1
else
  echo "Avvio mqtt-broker..."
  docker-compose -f "$COMPOSE_MQTT_BROKER_FILE_NAME" --env-file ".env" -p mqtt-broker up -d
fi

echo "docker-compose avviato con successo."

# Scarica il template del servizio
SERVICES_DIR="services"
SERVICE_FILE_NAME="sc-deploy.service"
TEMPLATE_SERVICE_FILE_NAME="$SERVICE_FILE_NAME.template"

# Controlla se il file di servizio esiste già
if [ -f "/etc/systemd/system/$SERVICE_FILE_NAME" ]; then
  echo "Il file di servizio /etc/systemd/system/$SERVICE_FILE_NAME esiste già. Esco."
  echo "Abilito il servizio all'avvio del sistema..."
  sudo systemctl enable "$SERVICE_FILE_NAME"
  exit 0
fi

# Scarica il file di template del servizio
echo "Scarico $TEMPLATE_SERVICE_FILE da s3://$BUCKET_NAME/$SERVICES_DIR/$TEMPLATE_SERVICE_FILE_NAME..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$SERVICES_DIR/$TEMPLATE_SERVICE_FILE_NAME" "$TEMPLATE_SERVICE_FILE_NAME"
if [ $? -ne 0 ]; then
  echo "Errore nel download di $TEMPLATE_SERVICE_FILE"
  exit 1
fi

# Crea il file del servizio specifico
if [ ! -f "$TEMPLATE_SERVICE_FILE_NAME" ]; then
  echo "File $TEMPLATE_SERVICE_FILE_NAME non trovato, esco."
  exit 1
fi

SCRIPT="$(basename "$0")"
echo "Creo il file di servizio /etc/systemd/system/$SERVICE_FILE_NAME..."
# Sostituisci il placeholder nel template e crea il file di servizio
echo "Sostituisco il placeholder \$SCRIPT con $SCRIPT..."
sudo sed "s|$SCRIPT|${SCRIPT}|g" \
  "$TEMPLATE_SERVICE_FILE_NAME" | sudo tee "/etc/systemd/system/$SERVICE_FILE_NAME" > /dev/null
echo "File di servizio creato in /etc/systemd/system/$SERVICE_FILE_NAME"

# Ricarica i file di configurazione di systemd
echo "Ricarico i file di configurazione di systemd..."
sudo systemctl daemon-reload

# Abilita il servizio all'avvio del sistema
echo "Abilito il servizio $SERVICE_FILE_NAME all'avvio del sistema..."
sudo systemctl enable "$SERVICE_FILE_NAME"

# Avvia il servizio
echo "Servizio $SERVICE_FILE_NAME creato e avviato."