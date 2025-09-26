#!/bin/bash

set -e

AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
COMPOSE_DIR="compose"
COMPOSE_EDGE_HUB_FILE_NAME="edge-hub.yaml"
COMPOSE_SENSORS_FILE_NAME="sensor-agent.generated_${SENSORS_NUMBER:-50}.yml"
DELAY_FILENAME="init-delay.sh"
ANALYZE_FILENAME="analyze_failure.sh"

ENDPOINT_URL=""
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "[INFO] Deploy su LocalStack..."
else
  echo "[INFO] Deploy su AWS..."
fi

echo "[DEBUG] Carico variabili d'ambiente da .env"
if [ -f ".env" ]; then
  source .env
  echo "[INFO] Variabili d'ambiente caricate da .env"
else
  echo "[ERROR] File .env non trovato!"
  exit 1
fi

echo "[DEBUG] Scarico $COMPOSE_EDGE_HUB_FILE_NAME da S3..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_EDGE_HUB_FILE_NAME" "$COMPOSE_EDGE_HUB_FILE_NAME"
if [ $? -ne 0 ]; then
  echo "[ERROR] Errore nel download di $COMPOSE_EDGE_HUB_FILE_NAME"
  exit 1
fi

echo "[DEBUG] Scarico $COMPOSE_SENSORS_FILE_NAME da S3..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_SENSORS_FILE_NAME" "$COMPOSE_SENSORS_FILE_NAME"
if [ $? -ne 0 ]; then
  echo "[ERROR] Errore nel download di $COMPOSE_SENSORS_FILE_NAME"
  exit 1
fi

echo "[DEBUG] Controllo volume zone-hub-${EDGE_ZONE}-cache-data..."
if [ ! "$(docker volume ls -q -f name=zone-hub-${EDGE_ZONE}-cache-data)" ]; then
  echo "[WARN] Volume zone-hub-${EDGE_ZONE}-cache-data non trovato. Creo il volume..."
  docker volume create zone-hub-${EDGE_ZONE}-cache-data || { echo "[ERROR] Errore nella creazione del volume"; exit 1; }
else
  echo "[INFO] Volume zone-hub-${EDGE_ZONE}-cache-data già esistente."
fi

if [ -f "$COMPOSE_EDGE_HUB_FILE_NAME" ]; then
  echo "[INFO] Elimino eventuali container edge-hub esistenti..."
  docker-compose -f "$COMPOSE_EDGE_HUB_FILE_NAME" --env-file ".env" -p edge-hub down
  echo "[INFO] Avvio edge-hub..."
  docker-compose -f "$COMPOSE_EDGE_HUB_FILE_NAME" --env-file ".env" -p edge-hub up -d
else
  echo "[ERROR] File $COMPOSE_EDGE_HUB_FILE_NAME non trovato!"
  exit 1
fi

if [ -f "$COMPOSE_SENSORS_FILE_NAME" ]; then
  echo "[INFO] Elimino eventuali container sensors esistenti..."
  docker-compose -f "$COMPOSE_SENSORS_FILE_NAME" --env-file ".env" -p sensors down
  echo "[INFO] Avvio sensori..."
  docker-compose -f "$COMPOSE_SENSORS_FILE_NAME" --env-file ".env" -p sensors up -d
else
  echo "[ERROR] File $COMPOSE_SENSORS_FILE_NAME non trovato!"
  exit 1
fi

CRON_JOB="0 3 * * * docker-compose -p sensors restart >> /var/log/sensors-restart.log 2>&1"
echo "[DEBUG] Verifico presenza job cron per riavvio sensori..."
( crontab -l 2>/dev/null | grep -F "$CRON_JOB" ) || \
( crontab -l 2>/dev/null; echo "$CRON_JOB" ) | crontab -
echo "[INFO] Job cron per riavvio sensori configurato."

echo "[DEBUG] Scarico $DELAY_FILENAME da S3..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/init/$DELAY_FILENAME" "$DELAY_FILENAME"
if [ $? -ne 0 ]; then
  echo "[ERROR] Errore nel download di $DELAY_FILENAME"
  exit 1
fi

echo "[DEBUG] Rendo eseguibile $DELAY_FILENAME"
chmod +x "$DELAY_FILENAME"

echo "[INFO] Applico latenza di rete all'istanza..."
sudo ./"$DELAY_FILENAME" apply --delay "${NETWORK_DELAY:-200ms}" --jitter "${NETWORK_JITTER:-50ms}" --loss "${NETWORK_LOSS:-5%}"

echo "[INFO] Zona avviata con successo."

# -----------------------------
# Configurazione servizio systemd
# -----------------------------
SERVICES_DIR="services"
SERVICE_FILE_NAME="sc-deploy.service"
TEMPLATE_SERVICE_FILE_NAME="$SERVICE_FILE_NAME.template"

if [ -f "/etc/systemd/system/$SERVICE_FILE_NAME" ]; then
  echo "[WARN] File di servizio già presente: /etc/systemd/system/$SERVICE_FILE_NAME"
  echo "[INFO] Abilito il servizio all'avvio del sistema..."
  sudo systemctl enable "$SERVICE_FILE_NAME"
else
  echo "[DEBUG] Scarico $TEMPLATE_SERVICE_FILE_NAME da S3..."
  aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$SERVICES_DIR/$TEMPLATE_SERVICE_FILE_NAME" "$TEMPLATE_SERVICE_FILE_NAME"
  if [ $? -ne 0 ]; then
    echo "[ERROR] Errore nel download di $TEMPLATE_SERVICE_FILE_NAME"
    exit 1
  fi

  if [ ! -f "$TEMPLATE_SERVICE_FILE_NAME" ]; then
    echo "[ERROR] File $TEMPLATE_SERVICE_FILE_NAME non trovato!"
    exit 1
  fi

  SCRIPT="$(basename "$0")"
  echo "[DEBUG] Creo il file di servizio systemd $SERVICE_FILE_NAME..."
  sudo sed "s|\$SCRIPT|${SCRIPT}|g" \
    "$TEMPLATE_SERVICE_FILE_NAME" | sudo tee "/etc/systemd/system/$SERVICE_FILE_NAME" > /dev/null
  echo "[INFO] File di servizio creato: /etc/systemd/system/$SERVICE_FILE_NAME"

  echo "[DEBUG] Ricarico configurazione systemd..."
  sudo systemctl daemon-reload

  echo "[INFO] Abilito il servizio $SERVICE_FILE_NAME all'avvio del sistema..."
  sudo systemctl enable "$SERVICE_FILE_NAME"

  echo "[INFO] Servizio $SERVICE_FILE_NAME creato e avviato con successo."
fi

# -----------------------------
# Scarico script di analisi failure
# -----------------------------

echo "[DEBUG] Scarico $ANALYZE_FILENAME da S3..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/performance/$ANALYZE_FILENAME" "$ANALYZE_FILENAME"
if [ $? -ne 0 ]; then
  echo "[ERROR] Errore nel download di $ANALYZE_FILENAME"
  exit 1
fi

echo "[DEBUG] Rendo eseguibile $ANALYZE_FILENAME"
chmod +x "$ANALYZE_FILENAME"

echo "[INFO] Script di analisi failure scaricato con successo."