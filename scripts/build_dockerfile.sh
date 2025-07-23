#!/bin/bash

set -euo pipefail

# === CONTROLLA PARAMETRI: IMMAGINE e VERSIONE ===
if [ $# -lt 2 ]; then
    echo "[USO] ./build_sensor_agent.sh <immagine> <versione>"
    echo "Esempio: ./build_sensor_agent.sh sensor-agent 1.0.0"
    exit 1
fi

DOCKERFILE_NAME="$1"
IMAGE_NAME="$2"
VERSION="$3"
TARGET_DIR="SensorContinuum"

# === PERCORSO CORRENTE ===
CUR_DIR="$(pwd)"

# === VERIFICA PRESENZA DELLA CARTELLA SENSORCONTINUUM NEL PATH ===
if [[ "$CUR_DIR" != *"/$TARGET_DIR"* ]]; then
    echo "[ERRORE] La cartella \"$TARGET_DIR\" non è nel percorso corrente: $CUR_DIR"
    exit 1
fi

# === RISALITA AL ROOT DEL PROGETTO ===
FULL_PATH="$CUR_DIR"
while [[ "$(basename "$FULL_PATH")" != "$TARGET_DIR" ]]; do
    FULL_PATH="$(dirname "$FULL_PATH")"
    if [[ "$FULL_PATH" == "/" ]]; then
        echo "[ERRORE] Impossibile trovare la radice del progetto $TARGET_DIR."
        exit 1
    fi
done

echo "[INFO] Trovata radice del progetto: $FULL_PATH"
cd "$FULL_PATH"

# === COSTRUISCI PATH DEL DOCKERFILE ===
DOCKERFILE_PATH="deploy/docker/${DOCKERFILE_NAME}.Dockerfile"

# === VERIFICA ESISTENZA DOCKERFILE ===
if [ ! -f "$DOCKERFILE_PATH" ]; then
    echo "[ERRORE] Dockerfile non trovato: $DOCKERFILE_PATH"
    exit 1
fi

# === BUILD DOCKER ===
echo "[INFO] Avvio build Docker con tag: ${IMAGE_NAME}:${VERSION}"
docker build -f "$DOCKERFILE_PATH" -t "${IMAGE_NAME}:${VERSION}" .

echo "[OK] Build completata con successo: ${IMAGE_NAME}:${VERSION}"