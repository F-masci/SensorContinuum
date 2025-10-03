#!/bin/bash

# Importa le funzioni
source utils.sh

show_help() {
  echo "Utilizzo: $0 region-name [opzioni]"
  echo "  --deploy=localstack      Deploy su LocalStack invece che AWS"
  echo "  --aws-region REGION      Regione AWS (default: us-east-1)"
  echo "  --component COMPONENT      Componente da deployare (default: tutti)"
  echo "  --instance-type TYPE     Tipo di istanza EC2 (default: t3.small)"
  echo "  -h, --help               Mostra questo messaggio"
  echo "Esempio:"
  echo "  $0 region-001 --aws-region eu-east-1"
}

# Definizione dei template da usare
VPC_TEMPLATE="../cloudformation/region/VPC.yaml"
KAFKA_TEMPLATE="../cloudformation/region/Kafka.yaml"
DATABASES_TEMPLATE="../cloudformation/region/Databases.yaml"
SERVICES_TEMPLATE="../cloudformation/region/services.yaml"
DEPLOY_MODE="aws"
AWS_REGION="us-east-1"
COMPONENT="all"
INSTANCE_TYPE="t3.small"

REGION="$1"
if [[ -z "$REGION" ]]; then
  echo "Errore: il nome della regione è obbligatorio."
  show_help
  exit 1
fi
shift

# Parsing degli argomenti
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
    --component)
      COMPONENT="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

if [[ "$COMPONENT" != "all" && "$COMPONENT" != "vpc" && "$COMPONENT" != "kafka" && "$COMPONENT" != "databases" && "$COMPONENT" != "services" ]]; then
  echo "Errore: componente '$COMPONENT' non valido. Valori accettati: all, vpc, kafka, databases, services."
  exit 1
fi


# -----------------------------
# Deploy del template VPC
# -----------------------------

STACK_NAME="$REGION-stack"
VPC_NAME="$REGION-vpc"
SUBNET_NAME="$REGION-subnet"
VPC_CIDR="10.0.0.0/16"
SUBNET_CIDR="10.0.0.0/24"
SSH_CIDR="0.0.0.0/0"
SECURITY_GROUP_NAME="$REGION-sg"
HOSTED_ZONE_NAME="$REGION.sensor-continuum.local"
PUBLIC_HOSTED_ZONE_NAME="sensor-continuum.it"

# Imposta endpoint per LocalStack se richiesto
if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy su LocalStack..."
else
  ENDPOINT_URL=""
  echo "Deploy su AWS..."
fi

if [[ "$COMPONENT" == "all" || "$COMPONENT" == "vpc" ]]; then
  echo "Componente specificato: $COMPONENT. Eseguo il deploy del template VPC."

  # Deploy del template
  echo "Deploy del template VPC..."

  echo "Parametri usati:"
  echo "  Regione AWS: $AWS_REGION"
  echo "  Regione: $REGION"
  echo "  Nome Stack: $STACK_NAME"
  echo "  Nome VPC: $VPC_NAME"
  echo "  Nome Subnet: $SUBNET_NAME"
  echo "  CIDR VPC: $VPC_CIDR"
  echo "  CIDR Subnet: $SUBNET_CIDR"
  echo "  CIDR SSH: $SSH_CIDR"
  echo "  Security Group Name: $SECURITY_GROUP_NAME"
  echo "  Hosted Zone Name: $HOSTED_ZONE_NAME"

  aws cloudformation deploy \
    --template-file "$VPC_TEMPLATE" \
    --stack-name "$STACK_NAME" \
    --region "$AWS_REGION" \
    --capabilities CAPABILITY_NAMED_IAM \
    --parameter-overrides \
      VpcName="$VPC_NAME" \
      VpcCidr="$VPC_CIDR" \
      SubnetName="$SUBNET_NAME" \
      SubnetCidr="$SUBNET_CIDR" \
      SshCidr="$SSH_CIDR" \
      SecurityGroupName="$SECURITY_GROUP_NAME" \
      HostedZoneName="$HOSTED_ZONE_NAME" \
    $ENDPOINT_URL
  if [[ $? -ne 0 ]]; then
    echo "Errore nel deploy del template VPC."
    aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$STACK_NAME"
    exit 1
  fi
  echo "Deploy del template VPC completato con successo."

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

  echo "Verifica della VPC..."
  VPC_ID=$(aws ec2 $ENDPOINT_URL describe-vpcs --region "$AWS_REGION" --filters "Name=tag:Name,Values=$VPC_NAME" --query "Vpcs[0].VpcId" --output text)
  if [[ -z "$VPC_ID" || "$VPC_ID" == "None" ]]; then
    echo "Errore: VPC con tag Name=$VPC_NAME non trovata."
    exit 1
  fi
  echo "Trovato VPC ID: $VPC_ID per VPC $VPC_NAME"

  echo "Verifica della Subnet..."
  SUBNET_ID=$(aws ec2 $ENDPOINT_URL describe-subnets --region "$AWS_REGION" --filters "Name=tag:Name,Values=$SUBNET_NAME" --query "Subnets[0].SubnetId" --output text)
  if [[ -z "$SUBNET_ID" || "$SUBNET_ID" == "None" ]]; then
    echo "Errore: subnet con nome $SUBNET_NAME non trovata."
    exit 1
  fi
  echo "Trovata Subnet ID: $SUBNET_ID per Subnet $SUBNET_NAME"

  echo "Verifica del Security Group..."
  SECURITY_GROUP_ID=$(aws ec2 $ENDPOINT_URL describe-security-groups --region "$AWS_REGION" --filters "Name=tag:Name,Values=$SECURITY_GROUP_NAME" --query "SecurityGroups[0].GroupId" --output text)
  if [[ -z "$SECURITY_GROUP_ID" || "$SECURITY_GROUP_ID" == "None" ]]; then
    echo "Errore: Security Group con nome $SECURITY_GROUP_NAME non trovata."
    exit 1
  fi
  echo "Trovato Security Group ID: $SECURITY_GROUP_ID per Security Group $SECURITY_GROUP_NAME"

  echo "Verifica della Hosted Zone..."
  HOSTED_ZONE_ID=$(aws route53 $ENDPOINT_URL list-hosted-zones-by-name --dns-name "$HOSTED_ZONE_NAME" --query "HostedZones[0].Id" --output text | sed 's|/hostedzone/||')
  if [[ -z "$HOSTED_ZONE_ID" || "$HOSTED_ZONE_ID" == "None" ]]; then
    echo "Errore: Hosted Zone con nome $HOSTED_ZONE_NAME non trovata."
    exit 1
  fi
  echo "Trovata Hosted Zone ID: $HOSTED_ZONE_ID per Hosted Zone $HOSTED_ZONE_NAME"

  if [[ "$COMPONENT" == "vpc" ]]; then
    echo "Deploy completato per il componente vpc."
    exit 0
  fi

fi

# -----------------------------
# Deploy del template Kafka
# -----------------------------

KAFKA_STACK_NAME="$REGION-kafka-stack"
KAFKA_BROKER_NAME="$REGION-kafka-broker"
KAFKA_BROKER_HOSTNAME="kafka-broker.$HOSTED_ZONE_NAME"

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

# Creazione KeyPair per proximity services se necessario
KEYS_DIR="./keys"
KAFKA_KEY_NAME="$REGION-kafka-key"
KAFKA_KEY_FILE="$KEYS_DIR/$KAFKA_KEY_NAME.pem"

mkdir -p "$KEYS_DIR"

KAFKA_KEY_PAIR=$(
  { ensure_key_pair "$KAFKA_KEY_NAME" "$KAFKA_KEY_FILE" "$AWS_REGION" "$ENDPOINT_URL"; } | tee /dev/tty | tail -n 1
)

HOSTED_ZONE_ID=$(
  { find_hosted_zone_id "$HOSTED_ZONE_NAME" "$AWS_REGION" "$ENDPOINT_URL"; } | tee /dev/tty | tail -n 1
)

# Cerca un'AMI di Amazon Linux 2
IMAGE_ID=$(
  { find_amazon_linux_2_ami "$AWS_REGION" "$ENDPOINT_URL" "$DEPLOY_MODE"; } | tee /dev/tty | tail -n 1
)

# Cerca il file .env per la regione
ENV_FILE=$(
  { find_or_create_environment "$REGION"; } | tee /dev/tty | tail -n 1
)

if [[ "$COMPONENT" == "all" || "$COMPONENT" == "kafka" ]]; then
  echo "Componente specificato: $COMPONENT. Eseguo il deploy del template proximity services."

  echo "Deploy del template Kafka..."
  echo "Parametri usati:"
  echo "  Regione AWS: $AWS_REGION"
  echo "  Regione: $REGION"
  echo "  Nome Stack: $KAFKA_STACK_NAME"
  echo "  Tipo di istanza: $INSTANCE_TYPE"
  echo "  Nome VPC: $VPC_NAME"
  echo "  Nome Subnet: $SUBNET_NAME"
  echo "  Nome Security Group: $SECURITY_GROUP_NAME"
  echo "  Nome Hosted Zone: $HOSTED_ZONE_NAME"
  echo "  Nome Key Pair: $KAFKA_KEY_NAME"
  echo "  Nome istanza Kafka: $KAFKA_BROKER_NAME"
  echo "  Nome DNS Kafka: $KAFKA_BROKER_HOSTNAME"
  echo "  Environment File: $ENV_FILE"

  echo "Parametri calcolati:"
  echo "  VPC ID: $VPC_ID"
  echo "  Subnet ID: $SUBNET_ID"
  echo "  Security Group ID: $SECURITY_GROUP_ID"
  echo "  Hosted Zone ID: $HOSTED_ZONE_ID"
  echo "  AMI ID: $IMAGE_ID"

  aws cloudformation deploy \
  --template-file "$KAFKA_TEMPLATE" \
  --stack-name "$KAFKA_STACK_NAME" \
  --region "$AWS_REGION" \
  --capabilities CAPABILITY_NAMED_IAM \
  --parameter-overrides \
    InstanceType="$INSTANCE_TYPE" \
    ImageId="$IMAGE_ID" \
    SubnetId="$SUBNET_ID" \
    SecurityGroupId="$SECURITY_GROUP_ID" \
    HostedZoneId="$HOSTED_ZONE_ID" \
    KeyName="$KAFKA_KEY_NAME" \
    KafkaInstanceName="$KAFKA_BROKER_NAME" \
    KafkaInstanceHostname="$KAFKA_BROKER_HOSTNAME" \
    EnvironmentFile="$ENV_FILE" \
  $ENDPOINT_URL

  if [[ $? -ne 0 ]]; then
    echo "Errore nel deploy del template Kafka."
    aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$KAFKA_STACK_NAME"
    exit 1
  fi

  # Verifica lo stato dello stack
  echo "Verifica lo stato dello stack..."
  STACK_STATUS=$(aws cloudformation $ENDPOINT_URL describe-stacks --region "$AWS_REGION" --stack-name "$KAFKA_STACK_NAME" --query "Stacks[0].StackStatus" --output text)
  if [[ "$STACK_STATUS" != "CREATE_COMPLETE" && "$STACK_STATUS" != "UPDATE_COMPLETE" ]]; then
    echo "Errore: lo stack $KAFKA_STACK_NAME non è in stato CREATE_COMPLETE o UPDATE_COMPLETE. Stato attuale: $STACK_STATUS"
    aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$KAFKA_STACK_NAME"
    exit 1
  fi
  echo "Lo stack $KAFKA_STACK_NAME è in stato $STACK_STATUS."

  echo "Verifica la creazione delle risorse..."
  echo "Verifica della istanza Kafka..."
  KAFKA_INSTANCE_ID=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --filters "Name=tag:Name,Values=$KAFKA_BROKER_NAME" "Name=instance-state-name,Values=running,pending" --query "Reservations[0].Instances[0].InstanceId" --output text)
  if [[ -z "$KAFKA_INSTANCE_ID" || "$KAFKA_INSTANCE_ID" == "None" ]]; then
    echo "Errore: istanza Kafka con tag Name=$KAFKA_BROKER_NAME non trovata."
    exit 1
  fi
  echo "Trovata istanza Kafka ID: $KAFKA_INSTANCE_ID per istanza $KAFKA_BROKER_NAME"

  # Recupera IP privato e pubblico
  KAFKA_PRIVATE_IP=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --instance-ids "$KAFKA_INSTANCE_ID" --query "Reservations[0].Instances[0].PrivateIpAddress" --output text)
  KAFKA_PUBLIC_IP=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --instance-ids "$KAFKA_INSTANCE_ID" --query "Reservations[0].Instances[0].PublicIpAddress" --output text)

  echo " Kafka instance info:"
  echo "  IP Privato: $KAFKA_PRIVATE_IP"
  echo "  IP Pubblico: $KAFKA_PUBLIC_IP"
  echo "  Hostname: $KAFKA_BROKER_HOSTNAME"

  if [[ "$COMPONENT" == "kafka" ]]; then
    echo "Deploy completato per il componente kafka."
    exit 0
  fi

fi

# -----------------------------
# Deploy del template Databases
# -----------------------------

DATABASES_STACK_NAME="$REGION-databases-stack"
DATABASES_NAME="$REGION-databases"
SENSOR_DATABASE_HOSTNAME="measurement-db.$HOSTED_ZONE_NAME"
METADATA_DATABASE_HOSTNAME="metadata-db.$HOSTED_ZONE_NAME"

# Creazione KeyPair per proximity services se necessario
KEYS_DIR="./keys"
DATABASES_KEY_NAME="$REGION-databases-key"
DATABASES_KEY_FILE="$KEYS_DIR/$DATABASES_KEY_NAME.pem"

mkdir -p "$KEYS_DIR"

DATABASES_KEY_PAIR=$(
  { ensure_key_pair "$DATABASES_KEY_NAME" "$DATABASES_KEY_FILE" "$AWS_REGION" "$ENDPOINT_URL"; } | tee /dev/tty | tail -n 1
)

echo "Recupero ID della Hosted Zone $PUBLIC_HOSTED_ZONE_NAME..."
PUBLIC_HOSTED_ZONE_ID=$(aws route53 list-hosted-zones-by-name --dns-name "$PUBLIC_HOSTED_ZONE_NAME" --query "HostedZones[0].Id" --output text | sed 's|/hostedzone/||')
if [[ -z "$PUBLIC_HOSTED_ZONE_ID" || "$PUBLIC_HOSTED_ZONE_ID" == "None" ]]; then
  echo "Errore: Hosted Zone con nome $PUBLIC_HOSTED_ZONE_NAME non trovata."
  exit 1
fi
echo "Trovata Hosted Zone ID: $PUBLIC_HOSTED_ZONE_ID per Hosted Zone $PUBLIC_HOSTED_ZONE_NAME"

PUBLIC_REGION_DATABASES_HOSTNAME="${REGION}.databases.sensor-continuum.it"
PUBLIC_REGION_METADATA_DATABASES_HOSTNAME="${REGION}.metadata-db.sensor-continuum.it"
PUBLIC_REGION_SENSOR_DATABASES_HOSTNAME="${REGION}.measurement-db.sensor-continuum.it"

if [[ "$COMPONENT" == "all" || "$COMPONENT" == "databases" ]]; then
  echo "Componente specificato: $COMPONENT. Eseguo il deploy del template databases."

  echo "Deploy del template Databases..."
  echo "Parametri usati:"
  echo "  Regione AWS: $AWS_REGION"
  echo "  Regione: $REGION"
  echo "  Nome Stack: $DATABASES_STACK_NAME"
  echo "  Tipo di istanza: $INSTANCE_TYPE"
  echo "  Nome VPC: $VPC_NAME"
  echo "  Nome Subnet: $SUBNET_NAME"
  echo "  Nome Security Group: $SECURITY_GROUP_NAME"
  echo "  Nome Hosted Zone: $HOSTED_ZONE_NAME"
  echo "  Nome Hosted Zone Pubblica: $PUBLIC_HOSTED_ZONE_NAME"
  echo "  Nome Key Pair: $DATABASES_KEY_NAME"
  echo "  Nome istanza Databases: $DATABASES_NAME"
  echo "  Hostname Sensor Database: $SENSOR_DATABASE_HOSTNAME"
  echo "  Hostname Metadata Database: $METADATA_DATABASE_HOSTNAME"
  echo "  Public Hostname Sensor Database: $PUBLIC_REGION_SENSOR_DATABASES_HOSTNAME"
  echo "  Public Hostname Metadata Database: $PUBLIC_REGION_METADATA_DATABASES_HOSTNAME"
  echo "  Environment File: $ENV_FILE"

  echo "Parametri calcolati:"
  echo "  VPC ID: $VPC_ID"
  echo "  Subnet ID: $SUBNET_ID"
  echo "  Security Group ID: $SECURITY_GROUP_ID"
  echo "  Hosted Zone ID: $HOSTED_ZONE_ID"
  echo "  AMI ID: $IMAGE_ID"

  aws cloudformation deploy \
    --template-file "$DATABASES_TEMPLATE" \
    --stack-name "$DATABASES_STACK_NAME" \
    --region "$AWS_REGION" \
    --capabilities CAPABILITY_NAMED_IAM \
    --parameter-overrides \
      InstanceType="$INSTANCE_TYPE" \
      ImageId="$IMAGE_ID" \
      SubnetId="$SUBNET_ID" \
      SecurityGroupId="$SECURITY_GROUP_ID" \
      HostedZoneId="$HOSTED_ZONE_ID" \
      PublicHostedZoneId="$PUBLIC_HOSTED_ZONE_ID" \
      KeyName="$DATABASES_KEY_NAME" \
      RegionDatabasesInstanceName="$DATABASES_NAME" \
      SensorRegionDatabaseInstanceHostname="$SENSOR_DATABASE_HOSTNAME" \
      MetadataRegionDatabaseInstanceHostname="$METADATA_DATABASE_HOSTNAME" \
      PublicRegionDatabasesHostname="$PUBLIC_REGION_DATABASES_HOSTNAME" \
      PublicRegionMetadataDatabasesHostname="$PUBLIC_REGION_METADATA_DATABASES_HOSTNAME" \
      PublicRegionSensorDatabasesHostname="$PUBLIC_REGION_SENSOR_DATABASES_HOSTNAME" \
      EnvironmentFile="$ENV_FILE" \
    $ENDPOINT_URL

  if [[ $? -ne 0 ]]; then
    echo "Errore nel deploy del template Databases."
    aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$DATABASES_STACK_NAME"
    exit 1
  fi

  # Verifica lo stato dello stack
  echo "Verifica lo stato dello stack..."
  STACK_STATUS=$(aws cloudformation $ENDPOINT_URL describe-stacks --region "$AWS_REGION" --stack-name "$DATABASES_STACK_NAME" --query "Stacks[0].StackStatus" --output text)
  if [[ "$STACK_STATUS" != "CREATE_COMPLETE" && "$STACK_STATUS" != "UPDATE_COMPLETE" ]]; then
    echo "Errore: lo stack $DATABASES_STACK_NAME non è in stato CREATE_COMPLETE o UPDATE_COMPLETE. Stato attuale: $STACK_STATUS"
    aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$DATABASES_STACK_NAME"
    exit 1
  fi
  echo "Lo stack $DATABASES_STACK_NAME è in stato $STACK_STATUS."

  echo "Verifica la creazione delle risorse..."

  echo "Verifica della istanza Databases..."
  DATABASES_INSTANCE_ID=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --filters "Name=tag:Name,Values=$REGION-databases" "Name=instance-state-name,Values=running,pending" --query "Reservations[0].Instances[0].InstanceId" --output text)
  if [[ -z "$DATABASES_INSTANCE_ID" || "$DATABASES_INSTANCE_ID" == "None" ]]; then
    echo "Errore: istanza Databases con tag Name=$REGION-databases non trovata."
    exit 1
  fi
  echo "Trovata istanza Databases ID: $DATABASES_INSTANCE_ID per istanza $REGION-databases"

  # Recupera IP privato e pubblico
  DATABASES_PRIVATE_IP=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --instance-ids "$DATABASES_INSTANCE_ID" --query "Reservations[0].Instances[0].PrivateIpAddress" --output text)
  DATABASES_PUBLIC_IP=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --instance-ids "$DATABASES_INSTANCE_ID" --query "Reservations[0].Instances[0].PublicIpAddress" --output text)

  echo " Databases instance info:"
  echo "  IP Privato: $DATABASES_PRIVATE_IP"
  echo "  IP Pubblico: $DATABASES_PUBLIC_IP"
  echo "  Hostname Sensor Database: $SENSOR_DATABASE_HOSTNAME"
  echo "  Hostname Metadata Database: $METADATA_DATABASE_HOSTNAME"

  if [[ "$COMPONENT" == "databases" ]]; then
    echo "Deploy completato per il componente databases."
    exit 0
  fi

fi

# -----------------------------
# Deploy del template service
# -----------------------------

SERVICES_STACK_NAME="$REGION-services-stack"
SERVICES_NAME="$REGION-services"

# Creazione KeyPair per proximity services se necessario
KEYS_DIR="./keys"
SERVICES_KEY_NAME="$REGION-services-key"
SERVICES_KEY_FILE="$KEYS_DIR/$SERVICES_KEY_NAME.pem"

mkdir -p "$KEYS_DIR"

SERVICES_KEY_PAIR=$(
  { ensure_key_pair "$SERVICES_KEY_NAME" "$SERVICES_KEY_FILE" "$AWS_REGION" "$ENDPOINT_URL"; } | tee /dev/tty | tail -n 1
)

HOSTED_ZONE_ID=$(
  { find_hosted_zone_id "$HOSTED_ZONE_NAME" "$AWS_REGION" "$ENDPOINT_URL"; } | tee /dev/tty | tail -n 1
)

# Cerca un'AMI di Amazon Linux 2
IMAGE_ID=$(
  { find_amazon_linux_2_ami "$AWS_REGION" "$ENDPOINT_URL" "$DEPLOY_MODE"; } | tee /dev/tty | tail -n 1
)

# Cerca il file .env per la regione
ENV_FILE=$(
  { find_or_create_environment "$REGION"; } | tee /dev/tty | tail -n 1
)

if [[ "$COMPONENT" == "all" || "$COMPONENT" == "services" ]]; then
  echo "Componente specificato: $COMPONENT. Eseguo il deploy del template services."

  echo "Deploy del template Services..."
  echo "Parametri usati:"
  echo "  Regione AWS: $AWS_REGION"
  echo "  Regione: $REGION"
  echo "  Nome Stack: $SERVICES_STACK_NAME"
  echo "  Tipo di istanza: $INSTANCE_TYPE"
  echo "  Nome VPC: $VPC_NAME"
  echo "  Nome Subnet: $SUBNET_NAME"
  echo "  Nome Security Group: $SECURITY_GROUP_NAME"
  echo "  Nome Hosted Zone: $HOSTED_ZONE_NAME"
  echo "  Nome Key Pair: $SERVICES_KEY_NAME"
  echo "  Nome istanza Services: $SERVICES_NAME"
  echo "  Environment File: $ENV_FILE"

  echo "Parametri calcolati:"
  echo "  VPC ID: $VPC_ID"
  echo "  Subnet ID: $SUBNET_ID"
  echo "  Security Group ID: $SECURITY_GROUP_ID"
  echo "  Hosted Zone ID: $HOSTED_ZONE_ID"
  echo "  AMI ID: $IMAGE_ID"

  aws cloudformation deploy \
    --template-file "$SERVICES_TEMPLATE" \
    --stack-name "$SERVICES_STACK_NAME" \
    --region "$AWS_REGION" \
    --capabilities CAPABILITY_NAMED_IAM \
    --parameter-overrides \
      InstanceType="$INSTANCE_TYPE" \
      ImageId="$IMAGE_ID" \
      SubnetId="$SUBNET_ID" \
      SecurityGroupId="$SECURITY_GROUP_ID" \
      HostedZoneId="$HOSTED_ZONE_ID" \
      KeyName="$SERVICES_KEY_NAME" \
      ServicesInstanceName="$SERVICES_NAME" \
      EnvironmentFile="$ENV_FILE" \
    $ENDPOINT_URL

  if [[ $? -ne 0 ]]; then
    echo "Errore nel deploy del template Services."
    aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$SERVICES_STACK_NAME"
    exit 1
  fi

  # Verifica lo stato dello stack
  echo "Verifica lo stato dello stack..."
  STACK_STATUS=$(aws cloudformation $ENDPOINT_URL describe-stacks --region "$AWS_REGION" --stack-name "$SERVICES_STACK_NAME" --query "Stacks[0].StackStatus" --output text)
  if [[ "$STACK_STATUS" != "CREATE_COMPLETE" && "$STACK_STATUS" != "UPDATE_COMPLETE" ]]; then
    echo "Errore: lo stack $SERVICES_STACK_NAME non è in stato CREATE_COMPLETE o UPDATE_COMPLETE. Stato attuale: $STACK_STATUS"
    aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$SERVICES_STACK_NAME"
    exit 1
  fi
  echo "Lo stack $SERVICES_STACK_NAME è in stato $STACK_STATUS."

  echo "Verifica la creazione delle risorse..."

  echo "Verifica della istanza Services..."
  SERVICES_INSTANCE_ID=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --filters "Name=tag:Name,Values=$REGION-services" "Name=instance-state-name,Values=running,pending" --query "Reservations[0].Instances[0].InstanceId" --output text)
  if [[ -z "$SERVICES_INSTANCE_ID" || "$SERVICES_INSTANCE_ID" == "None" ]]; then
    echo "Errore: istanza Services con tag Name=$REGION-services non trovata."
    exit 1
  fi
  echo "Trovata istanza Services ID: $SERVICES_INSTANCE_ID per istanza $REGION-services"

  # Recupera IP privato e pubblico
  SERVICES_PRIVATE_IP=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --instance-ids "$SERVICES_INSTANCE_ID" --query "Reservations[0].Instances[0].PrivateIpAddress" --output text)
  SERVICES_PUBLIC_IP=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --instance-ids "$SERVICES_INSTANCE_ID" --query "Reservations[0].Instances[0].PublicIpAddress" --output text)

  echo " Services instance info:"
  echo "  IP Privato: $SERVICES_PRIVATE_IP"
  echo "  IP Pubblico: $SERVICES_PUBLIC_IP"

  if [[ "$COMPONENT" == "services" ]]; then
    echo "Deploy completato per il componente services."
    exit 0
  fi

fi

echo "Deploy completato."