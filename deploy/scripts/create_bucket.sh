#!/bin/bash

set -e

AWS_REGION="${AWS_REGION:-us-east-1}"
BUCKET_NAME="${BUCKET_NAME:-sensor-continuum-scripts}"
DEPLOY_MODE="${DEPLOY_MODE:-aws}"
ENDPOINT_URL=""

if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy su LocalStack..."
else
  echo "Deploy su AWS..."
fi

declare -A FILE_MAP
FILE_MAP["inits/*install*.sh"]="init/"
FILE_MAP["inits/*init*.sh"]="init/"
FILE_MAP["deploy/*deploy*.sh"]="deploy/"
FILE_MAP["../compose/*.y*ml"]="compose/"
FILE_MAP["services/*.service.template"]="services/"

# Gestione flag --no-create
NO_CREATE=0
declare -a SELECTED_FILES=()
for arg in "$@"; do
  if [[ "$arg" == "--no-create" ]]; then
    NO_CREATE=1
  else
    SELECTED_FILES+=("$arg")
  fi
done

echo "Verifica bucket $BUCKET_NAME..."
if [[ $NO_CREATE -eq 0 ]]; then
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
else
  echo "Salto creazione bucket (--no-create)."
fi

file_selezionato() {
  local f="$1"
  local base="$(basename "$f")"
  if [[ ${#SELECTED_FILES[@]} -eq 0 ]]; then
    return 0
  fi
  for sel in "${SELECTED_FILES[@]}"; do
    [[ "$base" == "$sel" ]] && return 0
  done
  return 1
}

for pattern in "${!FILE_MAP[@]}"; do
  dest_path="${FILE_MAP[$pattern]}"
  files=( $pattern )
  for f in "${files[@]}"; do
    if [[ -f "$f" ]] && file_selezionato "$f"; then
      aws s3 cp $ENDPOINT_URL "$f" "s3://$BUCKET_NAME/$dest_path"
      echo "Copiato $f → s3://$BUCKET_NAME/$dest_path"
    elif [[ -f "$f" ]]; then
      echo "File $f non selezionato, salto."
    else
      echo "File $f non trovato, salto."
    fi
  done
done

echo "Deploy dei file completato."