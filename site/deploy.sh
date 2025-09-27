#!/bin/bash
set -e

# ----------------
# Configurazioni
# ----------------

BUCKET_NAME="sensor-continuum-site"   # Bucket S3 per la build

# ----------------
# Upload build React su S3
# ----------------

echo "[INFO] Sincronizzazione della cartella build/ su S3://$BUCKET_NAME"
aws s3 sync build/ s3://$BUCKET_NAME --delete
echo "[INFO] Sincronizzazione completata"
echo

echo "[INFO] Deploy su S3 completato"
echo "[INFO] Accedere alla dashboard di Amplify per poter implementare gli aggiornamenti"
