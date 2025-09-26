#!/bin/bash

set -euo pipefail

# -----------------------------------
# Configurazioni
# -----------------------------------
AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"

COMPOSE_DIR="compose"
COMPOSE_PROXIMITY_HUB_FILE_NAME="proximity-fog-hub.yaml"
DELAY_FILENAME="init-delay.sh"

SERVICES_DIR="services"
SERVICE_FILE_NAME="sc-deploy-services.service"
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
  echo "[ERROR] File .env non trovato!"
  exit 1
fi

# -----------------------------------
# Scarica file docker-compose
# -----------------------------------
echo "[INFO] Scarico $COMPOSE_PROXIMITY_HUB_FILE_NAME da S3..."
if ! aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_PROXIMITY_HUB_FILE_NAME" "$COMPOSE_PROXIMITY_HUB_FILE_NAME"; then
  echo "[ERROR] Errore nel download di $COMPOSE_PROXIMITY_HUB_FILE_NAME"
  exit 1
fi

# -----------------------------------
# Controlla volume docker
# -----------------------------------
VOLUME_NAME="macrozone-hub-${EDGE_MACROZONE}-cache-data"
echo "[INFO] Controllo se esiste il volume $VOLUME_NAME..."
if [[ -z "$(docker volume ls -q -f name=$VOLUME_NAME)" ]]; then
  echo "[INFO] Volume $VOLUME_NAME non trovato. Creo il volume..."
  docker volume create "$VOLUME_NAME" || { echo "[ERROR] Errore nella creazione del volume"; exit 1; }
else
  echo "[INFO] Volume $VOLUME_NAME già esistente."
fi

# -----------------------------------
# Avvia docker-compose
# -----------------------------------
if [[ ! -f "$COMPOSE_PROXIMITY_HUB_FILE_NAME" ]]; then
  echo "[ERROR] File $COMPOSE_PROXIMITY_HUB_FILE_NAME non trovato, esco."
  exit 1
fi

echo "[INFO] Elimino eventuali container proximity-fog-hub esistenti..."
docker-compose -f "$COMPOSE_PROXIMITY_HUB_FILE_NAME" --env-file ".env" -p proximity-fog-hub down

echo "[INFO] Avvio proximity-fog-hub cache..."
docker-compose -f "$COMPOSE_PROXIMITY_HUB_FILE_NAME" --env-file ".env" -p proximity-fog-hub up -d macrozone-hub-postgres-cache

# -----------------------------------
# Attendi che Postgres sia pronto
# -----------------------------------
echo "[INFO] Attendo che Postgres sia pronto..."
until docker exec "macrozone-hub-${EDGE_MACROZONE}-cache" pg_isready -U admin -d sensorcontinuum; do
  echo "[DEBUG] Postgres non ancora pronto, attendo..."
  sleep 2
done

echo "[INFO] Configuro Postgres: aumento max_connections a 500..."
until docker exec "macrozone-hub-${EDGE_MACROZONE}-cache" psql -U admin -d sensorcontinuum -c "ALTER SYSTEM SET max_connections = 500;"; do
  echo "[DEBUG] Comando non riuscito, riprovo..."
  sleep 1
done

echo "[INFO] Riavvio proximity-fog-hub per applicare le modifiche..."
docker-compose -f "$COMPOSE_PROXIMITY_HUB_FILE_NAME" --env-file ".env" -p proximity-fog-hub up -d

# -----------------------------------
# Configura latenza di rete
# -----------------------------------
echo "[INFO] Scarico $DELAY_FILENAME da S3..."
if ! aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/init/$DELAY_FILENAME" "$DELAY_FILENAME"; then
  echo "[ERROR] Errore nel download di $DELAY_FILENAME"
  exit 1
fi

chmod +x "$DELAY_FILENAME"
echo "[INFO] Applico latenza di rete..."
sudo "./$DELAY_FILENAME" apply \
  --delay "${NETWORK_DELAY:-30ms}" \
  --jitter "${NETWORK_JITTER:-15ms}" \
  --loss "${NETWORK_LOSS:-2%}"

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

echo "[INFO] Scarico il template $TEMPLATE_SERVICE_FILE_NAME da S3..."
if ! aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$SERVICES_DIR/$TEMPLATE_SERVICE_FILE_NAME" "$TEMPLATE_SERVICE_FILE_NAME"; then
  echo "[ERROR] Errore nel download di $TEMPLATE_SERVICE_FILE_NAME"
  exit 1
fi

if [[ ! -f "$TEMPLATE_SERVICE_FILE_NAME" ]]; then
  echo "[ERROR] File $TEMPLATE_SERVICE_FILE_NAME non trovato, esco."
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
echo "[INFO] Ricarico configurazioni systemd..."
sudo systemctl daemon-reload

echo "[INFO] Abilito il servizio $SERVICE_FILE_NAME all'avvio del sistema..."
sudo systemctl enable "$SERVICE_FILE_NAME"

echo "[INFO] Servizio $SERVICE_FILE_NAME creato e abilitato."
