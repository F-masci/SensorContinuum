#!/bin/bash

BUCKET_NAME="sensor-continuum"
DEPLOY_MODE="aws"

# File fissi da caricare
FIXED_FILES=(../compose/* ../compose/envs/*)
EXTRA_FILES=()

for arg in "$@"; do
  if [[ "$arg" == "--deploy=localstack" ]]; then
    DEPLOY_MODE="localstack"
  elif [[ "$arg" != --* ]]; then
    EXTRA_FILES+=("$arg")
  fi
done

if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy su LocalStack..."
else
  ENDPOINT_URL=""
  echo "Deploy su AWS..."
fi

# Controlla se il bucket esiste
aws s3api head-bucket --bucket "$BUCKET_NAME" $ENDPOINT_URL 2>/dev/null
if [[ $? -ne 0 ]]; then
  echo "Bucket $BUCKET_NAME non esiste, lo creo..."
  aws s3api create-bucket --bucket "$BUCKET_NAME" --region "$AWS_REGION" $ENDPOINT_URL
  if [[ $? -ne 0 ]]; then
    echo "Errore nella creazione del bucket $BUCKET_NAME."
    exit 1
  fi
  echo "Bucket $BUCKET_NAME creato con successo."
else
  echo "Bucket $BUCKET_NAME già esistente, aggiorno i file..."
fi

# Carica i file fissi e quelli extra
ALL_FILES=("${FIXED_FILES[@]}" "${EXTRA_FILES[@]}")
for file in "${ALL_FILES[@]}"; do
  if [[ -f "$file" ]]; then
    aws s3 cp "$file" "s3://$BUCKET_NAME/" $ENDPOINT_URL
    echo "File $file caricato su $BUCKET_NAME."
  else
    echo "File $file non trovato, salto."
  fi
done

echo "Operazione completata."