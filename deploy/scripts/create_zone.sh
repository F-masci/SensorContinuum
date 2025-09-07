#!/bin/bash

# Importa le funzioni
source utils.sh

show_help() {
  echo "Utilizzo: $0 region-name macrozone-name zone-name [opzioni]"
  echo "  --deploy=localstack      Deploy su LocalStack invece che AWS"
  echo "  --aws-region REGION      Regione AWS (default: us-east-1)"
  echo "  --instance-type TYPE     Tipo di istanza EC2 (default: t3.micro)"
  echo "  -h, --help               Mostra questo messaggio"
  echo "Esempio:"
  echo "  $0 region-001 macrozone-001 zone-001 --aws-region us-east-1 --instance-type t3.micro"
}

TEMPLATE_FILE="../cloudformation/zone/services.yaml"
DEPLOY_MODE="aws"
AWS_REGION="us-east-1"
INSTANCE_TYPE="t3.micro"

REGION="$1"
if [[ -z "$REGION" ]]; then
  echo "Errore: il nome della regione è obbligatorio."
  show_help
  exit 1
fi
shift

MACROZONE="$1"
if [[ -z "$MACROZONE" ]]; then
  echo "Errore: il nome della macrozona è obbligatorio."
  show_help
  exit 1
fi
shift

ZONE="$1"
if [[ -z "$ZONE" ]]; then
  echo "Errore: il nome della zona è obbligatorio."
  show_help
  exit 1
fi
shift

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      show_help
      exit 0
      ;;
    --deploy=localstack)
      DEPLOY_MODE="localstack"
      shift
      ;;
    --aws-region)
      AWS_REGION="$2"
      shift 2
      ;;
    --instance-type)
      INSTANCE_TYPE="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

STACK_NAME="$REGION-$MACROZONE-$ZONE-services-stack"
VPC_NAME="$REGION-vpc"
SUBNET_NAME="$REGION-$MACROZONE-subnet"
SECURITY_GROUP_NAME="$REGION-$MACROZONE-sg"
HOSTED_ZONE_NAME="$REGION.sensor-continuum.local"
ZONE_MQTT_BROKER_HOSTNAME="$ZONE.$MACROZONE.mqtt-broker.$HOSTED_ZONE_NAME"
MACROZONE_MQTT_BROKER_HOSTNAME="$MACROZONE.mqtt-broker.$HOSTED_ZONE_NAME"
SENSOR_MQTT_BROKER_HOSTNAME="$ZONE.$MACROZONE.sensor.mqtt-broker.$HOSTED_ZONE_NAME"
HUB_MQTT_BROKER_HOSTNAME="$ZONE.$MACROZONE.hub.mqtt-broker.$HOSTED_ZONE_NAME"
ROUTE_TABLE_NAME="$REGION-vpc-public-rt"
KEY_NAME="$REGION-$MACROZONE-$ZONE-edge-key"
SERVICES_INSTANCE_NAME="$REGION-$MACROZONE-$ZONE-services"

if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy su LocalStack..."
else
  ENDPOINT_URL=""
  echo "Deploy su AWS..."
fi

VPC_ID=$(
  { find_vpc_id "$VPC_NAME" "$AWS_REGION" "$ENDPOINT_URL"; } | tee /dev/tty | tail -n 1
) || exit 1

# Cerca il Subnet ID tramite il tag Name
SUBNET_ID=$(
  { find_subnet_id "$SUBNET_NAME" "$VPC_ID" "$AWS_REGION" "$ENDPOINT_URL"; } | tee /dev/tty | tail -n 1
)

# Cerca il Security Group ID tramite il tag Name
SECURITY_GROUP_ID=$(
  { find_sg_id "$SECURITY_GROUP_NAME" "$VPC_ID" "$AWS_REGION" "$ENDPOINT_URL"; } | tee /dev/tty | tail -n 1
)

# Cerca il Route Table ID tramite il tag Name
ROUTE_TABLE_ID=$(
  { find_route_table_id "$ROUTE_TABLE_NAME" "$AWS_REGION" "$ENDPOINT_URL"; } | tee /dev/tty | tail -n 1
) || exit 1


# Crea la key pair solo se non esiste già il file .pem
KEYS_DIR="keys"
KEY_FILE="$KEYS_DIR/$KEY_NAME.pem"

mkdir -p "$KEYS_DIR"

KEY_PAIR=$(
  { ensure_key_pair "$KEY_NAME" "$KEY_FILE" "$ENDPOINT_URL"; } | tee /dev/tty | tail -n 1
)

HOSTED_ZONE_ID=$(
  { find_hosted_zone_id "$HOSTED_ZONE_NAME" "$AWS_REGION" "$ENDPOINT_URL"; } | tee /dev/tty | tail -n 1
)

# Cerca un'AMI di Amazon Linux 2
IMAGE_ID=$(
  { find_amazon_linux_2_ami "$AWS_REGION" "$ENDPOINT_URL" "$DEPLOY_MODE"; } | tee /dev/tty | tail -n 1
)

# Cerca il file .env per la zona
ENV_FILE=$(
  { find_or_create_environment "$REGION" "$MACROZONE" "$ZONE"; } | tee /dev/tty | tail -n 1
)

# Deploy del template
echo "Deploy del template Edge Hub..."

echo "Parametri usati:"
echo "  Regione AWS: $AWS_REGION"
echo "  Regione: $REGION"
echo "  Macrozona: $MACROZONE"
echo "  Zona: $ZONE"
echo "  Nome Stack: $STACK_NAME"
echo "  Tipo di istanza: $INSTANCE_TYPE"
echo "  Nome VPC: $VPC_NAME"
echo "  Nome Subnet: $SUBNET_NAME"
echo "  Nome Security Group: $SECURITY_GROUP_NAME"
echo "  Nome Hosted Zone: $HOSTED_ZONE_NAME"
echo "  Nome DNS Broker MQTT di zona: $ZONE_MQTT_BROKER_HOSTNAME"
echo "  Nome DNS Broker MQTT di macrozona: $MACROZONE_MQTT_BROKER_HOSTNAME"
echo "  Nome DNS Broker MQTT di sensore: $SENSOR_MQTT_BROKER_HOSTNAME"
echo "  Nome DNS Broker MQTT di hub: $HUB_MQTT_BROKER_HOSTNAME"
echo "  Nome Route Table: $ROUTE_TABLE_NAME"
echo "  Nome Key Pair: $KEY_NAME"
echo "  Nome istanza servizi: $SERVICES_INSTANCE_NAME"
echo "  Environment file: $ENV_FILE"

echo "Parametri calcolati:"
echo "  VPC ID: $VPC_ID"
echo "  Subnet ID: $SUBNET_ID"
echo "  Security Group ID: $SECURITY_GROUP_ID"
echo "  Hosted Zone ID: $HOSTED_ZONE_ID"
echo "  Route Table ID: $ROUTE_TABLE_ID"
echo "  AMI ID: $IMAGE_ID"

aws cloudformation deploy \
  --template-file "$TEMPLATE_FILE" \
  --stack-name "$STACK_NAME" \
  --region "$AWS_REGION" \
  --capabilities CAPABILITY_NAMED_IAM \
  --parameter-overrides \
    InstanceType="$INSTANCE_TYPE" \
    ImageId="$IMAGE_ID" \
    SubnetId="$SUBNET_ID" \
    SecurityGroupId="$SECURITY_GROUP_ID" \
    HostedZoneId="$HOSTED_ZONE_ID" \
    ZoneMqttBrokerHostname="$ZONE_MQTT_BROKER_HOSTNAME" \
    MacrozoneMqttBrokerHostname="$MACROZONE_MQTT_BROKER_HOSTNAME" \
    SensorMqttBrokerHostname="$SENSOR_MQTT_BROKER_HOSTNAME" \
    HubMqttBrokerHostname="$HUB_MQTT_BROKER_HOSTNAME" \
    KeyName="$KEY_NAME" \
    ServicesInstanceName="$SERVICES_INSTANCE_NAME" \
    RouteTableId="$ROUTE_TABLE_ID" \
    EnvironmentFile="$ENV_FILE" \
  $ENDPOINT_URL
if [[ $? -ne 0 ]]; then
  echo "Errore nel deploy dello stack edge-hub."
    aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$STACK_NAME"
  exit 1
fi
echo "Deploy del template edge-hub completato con successo."

# Verifica lo stato dello stack
echo "Verifica lo stato dello stack..."
STACK_STATUS=$(aws cloudformation $ENDPOINT_URL describe-stacks --region "$AWS_REGION" --stack-name "$STACK_NAME" --query "Stacks[0].StackStatus" --output text)
if [[ "$STACK_STATUS" != "CREATE_COMPLETE" && "$STACK_STATUS" != "UPDATE_COMPLETE" ]]; then
  echo "Errore: lo stack $STACK_NAME non è in stato CREATE_COMPLETE o UPDATE_COMPLETE. Stato attuale: $STACK_STATUS"
  aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$STACK_NAME"
  exit 1
fi
echo "Lo stack $STACK_NAME è in stato $STACK_STATUS."

# Verifica la creazione delle risorse
echo "Verifica la creazione delle risorse..."
INSTANCE_ID=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --filters "Name=tag:Name,Values=$SERVICES_INSTANCE_NAME" "Name=subnet-id,Values=$SUBNET_ID" "Name=instance-state-name,Values=pending,running,stopping,stopped" --query "Reservations[0].Instances[0].InstanceId" --output text)
if [[ -z "$INSTANCE_ID" || "$INSTANCE_ID" == "None" ]]; then
  echo "Errore: istanza EC2 con tag Name=$SERVICES_INSTANCE_NAME non trovata nella Subnet $SUBNET_ID."
  exit 1
fi
echo "Trovata istanza EC2 ID: $INSTANCE_ID per i servizi $SERVICES_INSTANCE_NAME"

# Recupera IP privato e pubblico
SERVICES_INSTANCE_PRIVATE_IP=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --instance-ids "$INSTANCE_ID" --query "Reservations[0].Instances[0].PrivateIpAddress" --output text)
SERVICES_INSTANCE_PUBLIC_IP=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --instance-ids "$INSTANCE_ID" --query "Reservations[0].Instances[0].PublicIpAddress" --output text)

echo "Proximity services info:"
echo "  Private IP: $SERVICES_INSTANCE_PRIVATE_IP"
echo "  Public IP: $SERVICES_INSTANCE_PUBLIC_IP"

echo "Deploy completato."