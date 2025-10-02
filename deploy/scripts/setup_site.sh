#!/bin/bash

# ----------------
# Configurazioni
# ----------------

APP_NAME="Sensor Continuum"           # Nome dell'app Amplify
REGION="us-east-1"                    # Regione AWS
BRANCH_NAME="main"                    # Branch collegato all'app Amplify
CUSTOM_DOMAIN="sensor-continuum.it"   # Dominio personalizzato da associare

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
# Controllo e creazione branch Amplify
# ----------------

# Controlla se esiste già il branch; se no, lo crea
echo "[INFO] Controllo esistenza branch: $BRANCH_NAME"
BRANCH_EXISTS=$(aws amplify list-branches --app-id "$APP_ID" --query "branches[?branchName=='$BRANCH_NAME'].branchName" --output text)

if [ -z "$BRANCH_EXISTS" ] || [ "$BRANCH_EXISTS" == "None" ]; then
    echo "[INFO] Branch $BRANCH_NAME non trovato. Creazione branch..."
    aws amplify create-branch --app-id "$APP_ID" --branch-name "$BRANCH_NAME"
    echo "[INFO] Branch creato"
fi

# ----------------
# Associazione dominio personalizzato Amplify
# ----------------

echo "[INFO] Controllo associazione dominio personalizzato: $CUSTOM_DOMAIN"
DOMAIN_ASSOCIATION=$(aws amplify list-domain-associations --app-id "$APP_ID" --query "domainAssociations[?domainName=='$CUSTOM_DOMAIN'].domainName" --output text)

if [ -z "$DOMAIN_ASSOCIATION" ] || [ "$DOMAIN_ASSOCIATION" == "None" ]; then
    echo "[INFO] Dominio non associato. Avvio associazione dominio $CUSTOM_DOMAIN..."
    aws amplify create-domain-association \
        --app-id "$APP_ID" \
        --domain-name "$CUSTOM_DOMAIN" \
        --sub-domain-settings prefix=www,branchName="$BRANCH_NAME"
    echo "[INFO] Associazione dominio avviata. Controlla la console Amplify per i dettagli e la configurazione DNS."
else
    echo "[INFO] Dominio già associato: $CUSTOM_DOMAIN"
fi
echo

# ----------------
# Creazione bucket S3 se non esiste
# ----------------

BUCKET_NAME="sensor-continuum-site"
echo "[INFO] Verifica bucket S3: $BUCKET_NAME"
if ! aws s3 ls "s3://$BUCKET_NAME" 2>&1 | grep -q 'NoSuchBucket'; then
    echo "[INFO] Bucket già esistente."
else
    echo "[INFO] Bucket non trovato. Creazione..."
    aws s3 mb "s3://$BUCKET_NAME" --region "$REGION"
fi
echo

echo "[INFO] Setup completato. Accedi alla console Amplify per gestire l'app."
