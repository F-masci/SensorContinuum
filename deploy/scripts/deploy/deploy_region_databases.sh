#!/bin/bash

set -e

# Configurazioni
AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}" # "aws" o "localstack"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
COMPOSE_DIR="compose"
COMPOSE_REGION_DATABASES_FILE_NAME="region-databases.yml"

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
echo "Scarico $COMPOSE_REGION_DATABASES_FILE_NAME da s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_REGION_DATABASES_FILE_NAME..."
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_REGION_DATABASES_FILE_NAME" "$COMPOSE_REGION_DATABASES_FILE_NAME"
if [ $? -ne 0 ]; then
  echo "Errore nel download di $COMPOSE_REGION_DATABASES_FILE_NAME"
  exit 1
fi

# Avvia region-databases
if [ ! -f "$COMPOSE_REGION_DATABASES_FILE_NAME" ]; then
  echo "File $COMPOSE_REGION_DATABASES_FILE_NAME non trovato, esco."
  exit 1
fi

echo "Controllo se esiste il volume sensor-db-data-${REGION}..."
if [ ! "$(docker volume ls -q -f name=sensor-db-data-${REGION})" ]; then
  echo "Volume sensor-db-data-${REGION} non trovato. Creo il volume..."
  docker volume create sensor-db-data-${REGION} || { echo "Errore nella creazione del volume"; exit 1; }
else
  echo "Volume sensor-db-data-${REGION} già esistente."
fi

echo "Controllo se esiste il volume metadata-db-data-${REGION}..."
if [ ! "$(docker volume ls -q -f name=metadata-db-data-${REGION})" ]; then
  echo "Volume metadata-db-data-${REGION} non trovato. Creo il volume..."
  docker volume create metadata-db-data-${REGION} || { echo "Errore nella creazione del volume"; exit 1; }
else
  echo "Volume metadata-db-data-${REGION} già esistente."
fi

echo "Avvio region-databases..."
docker-compose -f "/home/ec2-user/region-databases.yml" --env-file ".env" -p region-databases up -d

echo "[DEBUG] Attendo che Postgres sia pronto..."
for i in {1..20}; do
  docker exec -it region-${REGION}-sensor-db pg_isready -U admin -d sensorcontinuum && \
    docker exec -it region-${REGION}-metadata-db pg_isready -U admin -d sensorcontinuum && \
    break
  sleep 2
done

# Aumenta il numero massimo di connessioni di postgres
echo "Aumento il numero massimo di connessioni di postgres a 500..."
docker exec -it region-${REGION}-sensor-db psql -U admin -d sensorcontinuum -c "ALTER SYSTEM SET max_connections = 500;"
# Riavvia il container per applicare la modifica
echo "Riavvio il container region-${REGION}-sensor-db per applicare la modifica..."
docker restart region-${REGION}-sensor-db

# Aumenta il numero massimo di connessioni di postgres
echo "Aumento il numero massimo di connessioni di postgres a 500..."
docker exec -it region-${REGION}-metadata-db psql -U admin -d sensorcontinuum -c "ALTER SYSTEM SET max_connections = 500;"
# Riavvia il container per applicare la modifica
echo "Riavvio il container region-${REGION}-metadata-db per applicare la modifica..."
docker restart region-${REGION}-metadata-db

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