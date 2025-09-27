#!/bin/bash
set -e

# ----------------
# Network setup
# ----------------

STACK_NAME="sc-lambda-network"                          # Nome dello stack CloudFormation
TEMPLATE_FILE="../cloudformation/lambda-network.yaml"   # Percorso al template yaml

echo "[INFO] Deploy stack $STACK_NAME con template $TEMPLATE_FILE"
aws cloudformation deploy \
  --stack-name "$STACK_NAME" \
  --template-file "$TEMPLATE_FILE" \
  --capabilities CAPABILITY_NAMED_IAM

echo "[INFO] Recupero output dello stack:"

# Recupera e mostra i valori utili per script lambda
VPC_ID=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='VPCId'].OutputValue" --output text)
SG_ID=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='LambdaSecurityGroupId'].OutputValue" --output text)
SUBNET1=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='PrivateSubnet1Id'].OutputValue" --output text)
SUBNET2=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='PrivateSubnet2Id'].OutputValue" --output text)

echo " VPC: $VPC_ID"
echo " SecurityGroup: $SG_ID"
echo " Subnet1: $SUBNET1"
echo " Subnet2: $SUBNET2"
echo

# ----------------
# API setup
# ----------------

STACK_NAME="sc-lambda-api"
TEMPLATE_FILE="../cloudformation/lambda-api.yaml"

echo "[INFO] Deploy stack $STACK_NAME con template $TEMPLATE_FILE"
aws cloudformation deploy \
  --stack-name "$STACK_NAME" \
  --template-file "$TEMPLATE_FILE"

echo "[INFO] Recupero output dello stack:"

API_ID=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='ApiId'].OutputValue" --output text)

echo " API ID: $API_ID"