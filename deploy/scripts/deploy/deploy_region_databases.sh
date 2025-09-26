#!/bin/bash

set -euo pipefail

# =============================
# Configurazioni
# =============================
AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}"   # "aws" o "localstack"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
COMPOSE_DIR="compose"
COMPOSE_FILE_NAME="region-databases.yml"
SERVICES_DIR="services"
SERVICE_FILE_NAME="sc-deploy.service"
TEMPLATE_SERVICE_FILE_NAME="$SERVICE_FILE_NAME.template"

# =============================
# Endpoint per LocalStack
# =============================
ENDPOINT_URL=""
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "[INFO] Deploy su LocalStack..."
else
  echo "[INFO] Deploy su AWS..."
fi

# =============================
# Carica variabili d'ambiente
# =============================
if [[ -f "./.env" ]]; then
  source ./.env
  echo "[INFO] Variabili d'ambiente caricate da .env"
else
  echo "[ERRORE] File .env non trovato!"
  exit 1
fi

# =============================
# Scarica il file docker-compose
# =============================
echo "[INFO] Scarico $COMPOSE_FILE_NAME da s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_FILE_NAME..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_FILE_NAME" "$COMPOSE_FILE_NAME" || {
  echo "[ERRORE] Download di $COMPOSE_FILE_NAME fallito."
  exit 1
}

[[ -f "$COMPOSE_FILE_NAME" ]] || { echo "[ERRORE] File $COMPOSE_FILE_NAME non trovato, esco."; exit 1; }

# =============================
# Gestione volumi Docker
# =============================
for VOL in "sensor-db-data-${REGION}" "metadata-db-data-${REGION}"; do
  echo "[INFO] Controllo volume $VOL..."
  if [[ -z "$(docker volume ls -q -f name=$VOL)" ]]; then
    echo "[INFO] Volume $VOL non trovato. Lo creo..."
    docker volume create "$VOL" || { echo "[ERRORE] Creazione volume $VOL fallita."; exit 1; }
  else
    echo "[INFO] Volume $VOL già esistente."
  fi
done

# =============================
# Avvia docker-compose
# =============================
echo "[INFO] Stop eventuali container region-databases..."
docker-compose -f "$COMPOSE_FILE_NAME" --env-file ".env" -p region-databases down || true

echo "[INFO] Avvio region-databases..."
docker-compose -f "$COMPOSE_FILE_NAME" --env-file ".env" -p region-databases up -d

# =============================
# Funzione: attendi Postgres pronto
# =============================
wait_for_postgres() {
  local container=$1
  local db=$2
  echo "[DEBUG] Attendo che Postgres in $container sia pronto..."
  until docker exec "$container" pg_isready -U admin -d "$db" >/dev/null 2>&1; do
    echo "[DEBUG] $container non pronto, attendo..."
    sleep 2
  done
}

# =============================
# Funzione: aumenta max_connections
# =============================
increase_max_connections() {
  local container=$1
  local db=$2
  echo "[INFO] Imposto max_connections=500 su $container..."
  until docker exec "$container" psql -U admin -d "$db" -c "ALTER SYSTEM SET max_connections = 500;" >/dev/null 2>&1; do
    echo "[DEBUG] Comando fallito su $container, riprovo..."
    sleep 1
  done
  echo "[INFO] Riavvio $container per applicare la modifica..."
  docker restart "$container"
}

# =============================
# Configura i database
# =============================
wait_for_postgres "region-${REGION}-sensor-db" "sensorcontinuum"
increase_max_connections "region-${REGION}-sensor-db" "sensorcontinuum"

wait_for_postgres "region-${REGION}-metadata-db" "sensorcontinuum"
increase_max_connections "region-${REGION}-metadata-db" "sensorcontinuum"

echo "[INFO] docker-compose avviato e database configurati con successo."

# =============================
# Configura systemd service
# =============================
if [[ -f "/etc/systemd/system/$SERVICE_FILE_NAME" ]]; then
  echo "[INFO] File di servizio già presente: /etc/systemd/system/$SERVICE_FILE_NAME"
  echo "[INFO] Abilito il servizio all'avvio del sistema..."
  sudo systemctl enable "$SERVICE_FILE_NAME"
  exit 0
fi

# Scarica il template
echo "[INFO] Scarico $TEMPLATE_SERVICE_FILE_NAME da s3://$BUCKET_NAME/$SERVICES_DIR/$TEMPLATE_SERVICE_FILE_NAME..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$SERVICES_DIR/$TEMPLATE_SERVICE_FILE_NAME" "$TEMPLATE_SERVICE_FILE_NAME" || {
  echo "[ERRORE] Download di $TEMPLATE_SERVICE_FILE_NAME fallito."
  exit 1
}

[[ -f "$TEMPLATE_SERVICE_FILE_NAME" ]] || { echo "[ERRORE] File template non trovato."; exit 1; }

# Genera il service file
SCRIPT="$(basename "$0")"
echo "[INFO] Creo il file di servizio /etc/systemd/system/$SERVICE_FILE_NAME..."
sudo sed "s|\$SCRIPT|${SCRIPT}|g" "$TEMPLATE_SERVICE_FILE_NAME" | sudo tee "/etc/systemd/system/$SERVICE_FILE_NAME" >/dev/null

# Ricarica e abilita il servizio
echo "[INFO] Ricarico configurazione systemd..."
sudo systemctl daemon-reload

echo "[INFO] Abilito il servizio $SERVICE_FILE_NAME all'avvio del sistema..."
sudo systemctl enable "$SERVICE_FILE_NAME"

echo "[INFO] Servizio $SERVICE_FILE_NAME creato e abilitato."
