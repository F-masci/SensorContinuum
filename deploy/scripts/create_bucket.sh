#!/bin/bash

set -e

# Configurazioni
AWS_REGION="${AWS_REGION:-us-east-1}"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}" # "aws" o "localstack"
ENDPOINT_URL=""

# Se deploy su localstack
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy su LocalStack..."
else
  echo "Deploy su AWS..."
fi

# Mapping locale -> path S3
# Formato: ["pattern locale"]="destinazione S3"
declare -A FILE_MAP
FILE_MAP["inits/*install*.sh"]="init/"
FILE_MAP["inits/*init*.sh"]="init/"
FILE_MAP["deploy/*deploy*.sh"]="deploy/"
FILE_MAP["../compose/*.y*ml"]="compose/"

echo "Verifica bucket $BUCKET_NAME..."
EXISTS=$(aws s3api $ENDPOINT_URL head-bucket --bucket "$BUCKET_NAME" 2>/dev/null || echo "no")
if [[ "$EXISTS" == "no" ]]; then
  echo "Bucket non trovato. Creazione bucket $BUCKET_NAME in $AWS_REGION..."
  aws s3api $ENDPOINT_URL create-bucket \
      --bucket "$BUCKET_NAME" \
      --region "$AWS_REGION" \
      $( [[ "$AWS_REGION" != "us-east-1" ]] && echo "--create-bucket-configuration LocationConstraint=$AWS_REGION" )
  echo "Bucket creato."
else
  echo "Bucket già esistente."
fi

# Copia file
for pattern in "${!FILE_MAP[@]}"; do
  dest_path="${FILE_MAP[$pattern]}"
  echo "Copia file matching '$pattern' in s3://$BUCKET_NAME/$dest_path"

  files=( $pattern )
  if [[ ${#files[@]} -eq 0 ]]; then
    echo "Attenzione: nessun file trovato per pattern '$pattern'"
    continue
  fi

  for f in "${files[@]}"; do
    if [[ -f "$f" ]]; then
      aws s3 cp $ENDPOINT_URL "$f" "s3://$BUCKET_NAME/$dest_path"
      echo "Copiato $f → s3://$BUCKET_NAME/$dest_path"
    else
      echo "File $f non trovato, salto."
    fi
  done
done

echo "Deploy dei file completato."
