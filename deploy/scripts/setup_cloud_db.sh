#!/bin/bash
set -e

STACK_NAME="sc-cloud-metadata-db"
TEMPLATE_FILE="../cloudformation/cloud-db.yaml"
PUBLIC_HOSTED_ZONE_NAME="sensor-continuum.it"

echo "[INFO] Deploy cluster..."
aws cloudformation deploy \
  --stack-name "$STACK_NAME" \
  --template-file "$TEMPLATE_FILE" \
  --parameter-overrides VpcId="$VPC_ID" Subnet1="$SUBNET1" Subnet2="$SUBNET2"

# Recupera Hosted Zone ID
HOSTED_ZONE_ID=$(aws route53 list-hosted-zones-by-name \
  --dns-name "$PUBLIC_HOSTED_ZONE_NAME" \
  --query "HostedZones[0].Id" \
  --output text)

# Route53 restituisce l'ID con prefisso /hostedzone/, rimuoviamolo
HOSTED_ZONE_ID=${HOSTED_ZONE_ID#/hostedzone/}

# Recupera endpoint writer e reader dal DB
WRITER_ENDPOINT=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" \
  --query "Stacks[0].Outputs[?OutputKey=='CloudMetaDbEndpoint'].OutputValue" --output text)
READER_ENDPOINT=$(aws cloudformation describe-stacks --stack-name "$STACK_NAME" \
  --query "Stacks[0].Outputs[?OutputKey=='CloudMetaDbReaderEndpoint'].OutputValue" --output text)

echo "[INFO] DB Writer Endpoint: $WRITER_ENDPOINT"
echo "[INFO] DB Reader Endpoint: $READER_ENDPOINT"

# Crea file JSON per i record
cat > /tmp/metadata-db-records.json <<EOF
{
  "Comment": "Aggiunta record per cloud.metadata-db.sensor-continuum.it",
  "Changes": [
    {
      "Action": "UPSERT",
      "ResourceRecordSet": {
        "Name": "write.cloud.metadata-db.sensor-continuum.it.",
        "Type": "CNAME",
        "TTL": 300,
        "ResourceRecords": [{ "Value": "$WRITER_ENDPOINT" }]
      }
    },
    {
      "Action": "UPSERT",
      "ResourceRecordSet": {
        "Name": "read-only.cloud.metadata-db.sensor-continuum.it.",
        "Type": "CNAME",
        "TTL": 300,
        "ResourceRecords": [{ "Value": "$READER_ENDPOINT" }]
      }
    },
    {
      "Action": "UPSERT",
      "ResourceRecordSet": {
        "Name": "cloud.metadata-db.sensor-continuum.it.",
        "Type": "CNAME",
        "TTL": 300,
        "ResourceRecords": [{ "Value": "read-only.cloud.metadata-db.sensor-continuum.it." }]
      }
    }
  ]
}
EOF

# Applica i record
aws route53 change-resource-record-sets --hosted-zone-id "$HOSTED_ZONE_ID" --change-batch file:///tmp/metadata-db-records.json

echo "[INFO] Record DNS creati nella zona sensor-continuum.it"

# Esegui gli script SQL sul database cloud dei metadati
PGHOST="$WRITER_ENDPOINT"
PGPORT=5433
PGUSER="sc_master"
PGPASSWORD="adminpass"
PGDATABASE="sensorcontinuum"

export PGPASSWORD

echo "[INFO] Eseguo init-cloud-metadata-db.sql..."
psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -f ../../configs/postgresql/init-cloud-metadata-db.sql

echo "[INFO] Eseguo init-metadata.sql..."
psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -f ../../configs/postgresql/init-metadata.sql

unset PGPASSWORD