#!/bin/bash
set -e

FOLDER=$1
FUNCTION=$2
PATH_ROUTE=$3       # percorso della route, es: /zone/sensor/data/raw/{region}/{macrozone}/{zone}/{sensor}
RESET=$5            # opzionale: --reset

API_ID=kuq3xgmhp3

if [ -z "$FOLDER" ] || [ -z "$FUNCTION" ] || [ -z "$PATH_ROUTE" ]; then
  echo "Usage: $0 <folder> <function> <path_route> <api_id> [--reset]"
  exit 1
fi

BASE_TEMPLATE="deploy/cloudformation/lambda.template.yaml"
OUTPUT_DIR="deploy/cloudformation/lambda/$FOLDER"
OUTPUT_FILE="$OUTPUT_DIR/$FUNCTION.yaml"
CONFIG_DIR=".aws-sam/configs/$FOLDER"
CONFIG_FILE="$CONFIG_DIR/$FUNCTION.toml"

mkdir -p "$OUTPUT_DIR"
mkdir -p "$CONFIG_DIR"

# --- Generazione template ---
echo "[INFO] Creo template per $FUNCTION"
sed \
  -e "s|\${FUNCTION_NAME}|$FUNCTION|g" \
  -e "s|\${FOLDER_PATH}|$FOLDER|g" \
  "$BASE_TEMPLATE" > "$OUTPUT_FILE"

# --- Build ---
echo "[INFO] Eseguo sam build"
sam build --template-file "$OUTPUT_FILE"

# --- Deploy ---
echo "[INFO] Avvio deploy con SAM"
sam deploy --template-file "$OUTPUT_FILE" --guided

# --- Aggiungi permessi Lambda per API Gateway ---
echo "[INFO] Aggiungo permessi Lambda per API Gateway"
aws lambda add-permission \
  --function-name "$FUNCTION" \
  --statement-id "apigw-invoke-$(date +%s)" \
  --action lambda:InvokeFunction \
  --principal apigateway.amazonaws.com \
  --source-arn "arn:aws:execute-api:us-east-1:975050105348:$API_ID/*/GET$PATH_ROUTE"

# --- Crea integrazione Lambda HTTP API ---
echo "[INFO] Creo integrazione HTTP API"
INTEGRATION_ID=$(aws apigatewayv2 create-integration \
  --api-id "$API_ID" \
  --integration-type AWS_PROXY \
  --integration-uri "arn:aws:lambda:us-east-1:975050105348:function:$FUNCTION" \
  --payload-format-version 2.0 \
  --request-parameters overwrite:path="\$request.path" \
  --query 'IntegrationId' --output text)

echo "[INFO] Integrazione creata con ID: $INTEGRATION_ID"

# --- Crea route GET sulla API con integrazione ---
echo "[INFO] Creo route GET $PATH_ROUTE"
aws apigatewayv2 create-route \
  --api-id "$API_ID" \
  --route-key "GET $PATH_ROUTE" \
  --target "integrations/$INTEGRATION_ID"

echo "[INFO] Lambda $FUNCTION agganciata a HTTP API $API_ID su route GET $PATH_ROUTE con PATH mapping $request.path"
