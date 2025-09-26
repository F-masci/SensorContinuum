#!/bin/bash

set -e

# -----------------------------
# Configurazioni
# -----------------------------
AWS_REGION="${AWS_REGION:-us-east-1}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}" # "aws" o "localstack"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
COMPOSE_DIR="compose"
COMPOSE_KAFKA_BROKER_FILE_NAME="kafka-broker.yml"

# -----------------------------
# Endpoint per LocalStack
# -----------------------------
ENDPOINT_URL=""
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "[INFO] Deploy su LocalStack..."
else
  echo "[INFO] Deploy su AWS..."
fi

# -----------------------------
# Carica variabili d'ambiente
# -----------------------------
echo "[DEBUG] Carico variabili d'ambiente da .env"
if [ -f "./.env" ]; then
  source ./.env
  echo "[INFO] Variabili d'ambiente caricate da .env"
else
  echo "[ERROR] File .env non trovato!"
  exit 1
fi

# -----------------------------
# Scarica il file compose
# -----------------------------
echo "[DEBUG] Scarico $COMPOSE_KAFKA_BROKER_FILE_NAME da S3"
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$COMPOSE_DIR/$COMPOSE_KAFKA_BROKER_FILE_NAME" "$COMPOSE_KAFKA_BROKER_FILE_NAME"
if [ $? -ne 0 ]; then
  echo "[ERROR] Errore nel download di $COMPOSE_KAFKA_BROKER_FILE_NAME"
  exit 1
fi

if [ ! -f "$COMPOSE_KAFKA_BROKER_FILE_NAME" ]; then
  echo "[ERROR] File $COMPOSE_KAFKA_BROKER_FILE_NAME non trovato, esco."
  exit 1
fi

# -----------------------------
# Volume Docker
# -----------------------------
echo "[DEBUG] Controllo se esiste il volume kafka-data-${REGION}"
if [ ! "$(docker volume ls -q -f name=kafka-data-${REGION})" ]; then
  echo "[INFO] Volume kafka-data-${REGION} non trovato. Creo il volume..."
  docker volume create kafka-data-${REGION} || { echo "[ERROR] Errore nella creazione del volume"; exit 1; }
else
  echo "[INFO] Volume kafka-data-${REGION} già esistente."
fi

# -----------------------------
# Configurazione IP pubblico
# -----------------------------
echo "[DEBUG] Recupero IP pubblico dell'istanza..."
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
sed -i "s/^KAFKA_PUBLIC_IP=.*/KAFKA_PUBLIC_IP=${PUBLIC_IP}/" .env
echo "[INFO] IP pubblico configurato: $PUBLIC_IP"

# -----------------------------
# Avvio Kafka broker
# -----------------------------
echo "[INFO] Elimino eventuali container kafka-broker esistenti..."
docker-compose -f "/home/ec2-user/$COMPOSE_KAFKA_BROKER_FILE_NAME" --env-file ".env" -p kafka-broker down
echo "[INFO] Avvio kafka-broker..."
docker-compose -f "/home/ec2-user/$COMPOSE_KAFKA_BROKER_FILE_NAME" --env-file ".env" -p kafka-broker up -d

echo "[INFO] docker-compose avviato con successo."

# -----------------------------
# Configurazione systemd
# -----------------------------
SERVICES_DIR="services"
SERVICE_FILE_NAME="sc-deploy.service"
TEMPLATE_SERVICE_FILE_NAME="$SERVICE_FILE_NAME.template"

echo "[DEBUG] Controllo se esiste già il servizio systemd"
if [ -f "/etc/systemd/system/$SERVICE_FILE_NAME" ]; then
  echo "[WARN] Il file di servizio /etc/systemd/system/$SERVICE_FILE_NAME esiste già."
  echo "[INFO] Abilito il servizio all'avvio del sistema..."
  sudo systemctl enable "$SERVICE_FILE_NAME"
  exit 0
fi

echo "[DEBUG] Scarico $TEMPLATE_SERVICE_FILE_NAME da S3"
aws s3 cp $ENDPOINT_URL "s3://$BUCKET_NAME/$SERVICES_DIR/$TEMPLATE_SERVICE_FILE_NAME" "$TEMPLATE_SERVICE_FILE_NAME"
if [ $? -ne 0 ]; then
  echo "[ERROR] Errore nel download di $TEMPLATE_SERVICE_FILE_NAME"
  exit 1
fi

if [ ! -f "$TEMPLATE_SERVICE_FILE_NAME" ]; then
  echo "[ERROR] File $TEMPLATE_SERVICE_FILE_NAME non trovato, esco."
  exit 1
fi

SCRIPT="$(basename "$0")"
echo "[INFO] Creo il file di servizio /etc/systemd/system/$SERVICE_FILE_NAME..."
echo "[DEBUG] Sostituisco il placeholder \$SCRIPT con $SCRIPT"
sudo sed "s|\$SCRIPT|${SCRIPT}|g" \
  "$TEMPLATE_SERVICE_FILE_NAME" | sudo tee "/etc/systemd/system/$SERVICE_FILE_NAME" > /dev/null
echo "[INFO] File di servizio creato in /etc/systemd/system/$SERVICE_FILE_NAME"

echo "[DEBUG] Ricarico configurazioni systemd..."
sudo systemctl daemon-reload

echo "[INFO] Abilito il servizio $SERVICE_FILE_NAME all'avvio..."
sudo systemctl enable "$SERVICE_FILE_NAME"

echo "[INFO] Servizio $SERVICE_FILE_NAME creato e avviato con successo."