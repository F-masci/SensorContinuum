#!/bin/bash
set -e

# ----------------
# Configurazioni
# ----------------

APP_NAME="sensor-continuum"           # Nome dell'app Amplify
BUCKET_NAME="sensor-continuum-site"   # Bucket S3 per la build
REGION="us-east-1"                    # Regione AWS
BRANCH_NAME="main"                    # Branch collegato all'app Amplify

# ----------------
# Controllo e creazione app Amplify
# ----------------

echo "[INFO] Controllo esistenza app Amplify: $APP_NAME"
APP_ID=$(aws amplify list-apps --query "apps[?name=='$APP_NAME'].appId" --output text)

if [ -z "$APP_ID" ] || [ "$APP_ID" == "None" ]; then
    echo "[INFO] App non trovata. Creazione app Amplify $APP_NAME..."
    APP_ID=$(aws amplify create-app --name "$APP_NAME" --region "$REGION" --query "app.appId" --output text)
    echo "[INFO] App Amplify creata con AppId: $APP_ID"
else
    echo "[INFO] App Amplify trovata: $APP_ID"
fi
echo

# ----------------
# Creazione bucket S3 se non esiste
# ----------------

echo "[INFO] Verifica bucket S3: $BUCKET_NAME"
if ! aws s3 ls "s3://$BUCKET_NAME" 2>&1 | grep -q 'NoSuchBucket'; then
    echo "[INFO] Bucket già esistente."
else
    echo "[INFO] Bucket non trovato. Creazione..."
    aws s3 mb "s3://$BUCKET_NAME" --region "$REGION"
fi
echo

# ----------------
# Upload build React su S3
# ----------------

echo "[INFO] Sincronizzazione della cartella build/ su S3://$BUCKET_NAME"
aws s3 sync build/ s3://$BUCKET_NAME --delete
echo "[INFO] Sincronizzazione completata"
echo

# ----------------
# Deploy su AWS Amplify
# ----------------

# Controlla se esiste già il branch; se no, lo crea
BRANCH_EXISTS=$(aws amplify list-branches --app-id "$APP_ID" --query "branches[?branchName=='$BRANCH_NAME'].branchName" --output text)
if [ -z "$BRANCH_EXISTS" ] || [ "$BRANCH_EXISTS" == "None" ]; then
    echo "[INFO] Branch $BRANCH_NAME non trovato. Creazione branch..."
    aws amplify create-branch --app-id "$APP_ID" --branch-name "$BRANCH_NAME"
    echo "[INFO] Branch creato"
fi

echo "[INFO] Deploy su S3 completato"
echo "[INFO] Accedere alla dashboard di Amplify per poter implementare gli aggiornamenti"
