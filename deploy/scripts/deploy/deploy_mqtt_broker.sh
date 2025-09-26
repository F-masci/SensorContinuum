#!/bin/bash

set -euo pipefail

# -----------------------------------
# Configurazioni
# -----------------------------------
AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}" # "aws" o "localstack"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"

COMPOSE_DIR="compose"
COMPOSE_MQTT_BROKER_FILE_NAME="mqtt-broker.yml"

SERVICES_DIR="services"
SERVICE_FILE_NAME="sc-deploy-mqtt.service"
TEMPLATE_SERVICE_FILE_NAME="sc-deploy.service.template"

# -----------------------------------
# Endpoint per LocalStack
# -----------------------------------
ENDPOINT_URL=""
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "[INFO] Deploy su LocalStack..."
else
  echo "[INFO] Deploy su AWS..."
fi

# -----------------------------------
# Carica variabili d'ambiente
# -----------------------------------
if [[ -f "./.env" ]]; then
  source ./.env
  echo "[INFO] Variabili d'ambiente caricate da .env"
else
  echo "[ERRORE] File .env non trovato!"
  exit 1
fi

# -----------------------------------
# Scarica il file docker-compose
# -----------------------------------
echo "[INFO] Scarico $COMPOSE_MQTT_BROKER_FILE_NAME da s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_MQTT_BROKER_FILE_NAME..."
if ! aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_MQTT_BROKER_FILE_NAME" "$COMPOSE_MQTT_BROKER_FILE_NAME"; then
  echo "[ERRORE] Download fallito: $COMPOSE_MQTT_BROKER_FILE_NAME"
  exit 1
fi

# -----------------------------------
# Avvia mqtt-broker
# -----------------------------------
if [[ ! -f "$COMPOSE_MQTT_BROKER_FILE_NAME" ]]; then
  echo "[ERRORE] File $COMPOSE_MQTT_BROKER_FILE_NAME non trovato, esco."
  exit 1
fi

echo "[INFO] Elimino eventuali container mqtt-broker esistenti..."
docker-compose -f "$COMPOSE_MQTT_BROKER_FILE_NAME" --env-file ".env" -p mqtt-broker down

echo "[INFO] Avvio mqtt-broker..."
docker-compose -f "$COMPOSE_MQTT_BROKER_FILE_NAME" --env-file ".env" -p mqtt-broker up -d

echo "[OK] docker-compose avviato con successo."

# -----------------------------------
# Configura systemd service
# -----------------------------------
if [[ -f "/etc/systemd/system/$SERVICE_FILE_NAME" ]]; then
  echo "[INFO] Il file di servizio $SERVICE_FILE_NAME esiste già."
  echo "[INFO] Abilito il servizio all'avvio del sistema..."
  sudo systemctl enable "$SERVICE_FILE_NAME"
  exit 0
fi

echo "[INFO] Scarico il template $TEMPLATE_SERVICE_FILE_NAME da s3://$BUCKET_NAME/$SERVICES_DIR/$TEMPLATE_SERVICE_FILE_NAME..."
if ! aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$SERVICES_DIR/$TEMPLATE_SERVICE_FILE_NAME" "$TEMPLATE_SERVICE_FILE_NAME"; then
  echo "[ERRORE] Download fallito: $TEMPLATE_SERVICE_FILE_NAME"
  exit 1
fi

if [[ ! -f "$TEMPLATE_SERVICE_FILE_NAME" ]]; then
  echo "[ERRORE] File $TEMPLATE_SERVICE_FILE_NAME non trovato, esco."
  exit 1
fi

SCRIPT="$(basename "$0")"

echo "[INFO] Creo il file di servizio /etc/systemd/system/$SERVICE_FILE_NAME..."
sudo sed "s|\$SCRIPT|${SCRIPT}|g" \
  "$TEMPLATE_SERVICE_FILE_NAME" | sudo tee "/etc/systemd/system/$SERVICE_FILE_NAME" > /dev/null

echo "[OK] File di servizio creato in /etc/systemd/system/$SERVICE_FILE_NAME"

# -----------------------------------
# Abilita e avvia il servizio
# -----------------------------------
echo "[INFO] Ricarico configurazioni di systemd..."
sudo systemctl daemon-reload

echo "[INFO] Abilito il servizio $SERVICE_FILE_NAME all'avvio del sistema..."
sudo systemctl enable "$SERVICE_FILE_NAME"

echo "[INFO] Servizio $SERVICE_FILE_NAME creato e abilitato."
