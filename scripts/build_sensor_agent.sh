#!/bin/bash

set -euo pipefail

# === CONTROLLA PARAMETRO VERSIONE ===
if [ $# -eq 0 ]; then
  echo "USO: $0 <versione>"
  echo "Esempio: $0 1.0.0"
  exit 1
fi
VERSION="$1"

# === PERCORSO CORRENTE ===
CUR_DIR="$(pwd)"
TARGET_DIR="SensorContinuum"

# === VERIFICA PRESENZA DELLA CARTELLA SENSORCONTINUUM ===
if [[ "$CUR_DIR" != *"$TARGET_DIR"* ]]; then
  echo "[ERRORE] La cartella \"$TARGET_DIR\" non è nel percorso corrente: $CUR_DIR"
  exit 1
fi

# === RISALITA AL ROOT DEL PROGETTO ===
FULL_PATH="$CUR_DIR"

while [ "$(basename "$FULL_PATH")" != "$TARGET_DIR" ]; do
  FULL_PATH="$(dirname "$FULL_PATH")"
  if [ "$FULL_PATH" == "/" ] || [ -z "$FULL_PATH" ]; then
    echo "[ERRORE] Impossibile trovare la radice del progetto."
    exit 1
  fi
done

echo "[INFO] Trovata radice del progetto: $FULL_PATH"
cd "$FULL_PATH"

# === BUILD DOCKER ===
echo "[INFO] Avvio build Docker con tag: sensor-agent:$VERSION"
docker build -f deploy/docker/sensor-agent.Dockerfile -t sensor-agent:"$VERSION" .

echo "[OK] Build completata con successo: sensor-agent:$VERSION"
