#!/bin/bash

set -e

AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
COMPOSE_DIR="compose"
COMPOSE_INTERMEDIATE_HUB_FILE_NAME="intermediate-fog-hub.yaml"
DELAY_FILENAME="init-delay.sh"

ENDPOINT_URL=""
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "[INFO] Deploy su LocalStack..."
else
  echo "[INFO] Deploy su AWS..."
fi

echo "[DEBUG] Carico variabili d'ambiente da .env"
if [ -f "./.env" ]; then
  source ./.env
  echo "[INFO] Variabili d'ambiente caricate da .env"
else
  echo "[ERROR] File .env non trovato!"
  exit 1
fi

echo "[DEBUG] Scarico $COMPOSE_INTERMEDIATE_HUB_FILE_NAME da S3"
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_INTERMEDIATE_HUB_FILE_NAME" "$COMPOSE_INTERMEDIATE_HUB_FILE_NAME"
if [ $? -ne 0 ]; then
  echo "[ERROR] Errore nel download di $COMPOSE_INTERMEDIATE_HUB_FILE_NAME"
  exit 1
fi

if [ ! -f "$COMPOSE_INTERMEDIATE_HUB_FILE_NAME" ]; then
  echo "[ERROR] File $COMPOSE_INTERMEDIATE_HUB_FILE_NAME non trovato, esco."
  exit 1
else
  echo "[INFO] Avvio intermediate-fog-hub..."
  docker-compose -f "$COMPOSE_INTERMEDIATE_HUB_FILE_NAME" --env-file ".env" -p intermediate-fog-hub up -d
fi

echo "[DEBUG] Scarico $DELAY_FILENAME da S3"
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/init/$DELAY_FILENAME" "$DELAY_FILENAME"
if [ $? -ne 0 ]; then
  echo "[ERROR] Errore nel download di $DELAY_FILENAME"
  exit 1
fi

echo "[DEBUG] Rendo eseguibile $DELAY_FILENAME"
chmod +x "$DELAY_FILENAME"

echo "[INFO] Applico latenza di rete all'istanza..."
sudo ./"$DELAY_FILENAME" apply --delay "${NETWORK_DELAY:-20ms}" --jitter "${NETWORK_JITTER:-5ms}"

echo "[INFO] docker-compose avviato con successo."

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