#!/bin/bash

STACK_NAME="sc-public-dns"
TEMPLATE_FILE="../cloudformation/public-dns.yaml"

echo "[INFO] Deploy stack $STACK_NAME con template $TEMPLATE_FILE"
aws cloudformation deploy \
  --stack-name "$STACK_NAME" \
  --template-file "$TEMPLATE_FILE"

echo "[INFO] Recupero Name Server:"
NS=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" \
  --query "Stacks[0].Outputs[?OutputKey=='NameServers'].OutputValue" --output text)
echo "[INFO] Name Server da configurare sul registrar:"
echo "$NS" | tr ',' '\n '
