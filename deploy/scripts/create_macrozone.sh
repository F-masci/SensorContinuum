#!/bin/bash

# Importa le funzioni
source utils.sh

show_help() {
  echo "Utilizzo: $0 region-name macrozone-name [opzioni]"
  echo "  --deploy=localstack        Deploy su LocalStack invece che AWS"
  echo "  --aws-region REGION        Regione AWS (default: eu-east-1)"
  echo "  --component COMPONENT      Componente da deployare (default: tutti)"
  echo "  --instance-type TYPE     Tipo di istanza EC2 (default: t2.micro)"
  echo "  -h, --help                 Mostra questo messaggio"
  echo "Esempio:"
  echo "  $0 region-001 macrozone-001 --aws-region us-east-1"
}

SUBNET_TEMPLATE="../terraform/macrozone/Subnet.yaml"
SERVICES_TEMPLATE="../terraform/macrozone/services.yaml"
DEPLOY_MODE="aws"
AWS_REGION="us-east-1"
COMPONENT="all"
INSTANCE_TYPE="t2.small"

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
    --component)
      COMPONENT="$2"
      echo "Componente da deployare: $COMPONENT"
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

if [[ "$COMPONENT" != "all" && "$COMPONENT" != "subnet" && "$COMPONENT" != "services" ]]; then
  echo "Errore: componente '$COMPONENT' non valido. Valori accettati: all, subnet, services."
  exit 1
fi



# -----------------------------
# Deploy del template Subnet
# -----------------------------

VPC_STACK_NAME="$REGION-$MACROZONE-stack"
VPC_NAME="$REGION-vpc"
SUBNET_NAME="$REGION-$MACROZONE-subnet"
SECURITY_GROUP_NAME="$REGION-$MACROZONE-sg"
ROUTE_TABLE_NAME="$REGION-vpc-public-rt"


if [[ "$DEPLOY_MODE" == "localstack" ]]; then
  ENDPOINT_URL="--endpoint-url=http://localhost:4566"
  echo "Deploy su LocalStack..."
else
  ENDPOINT_URL=""
  echo "Deploy su AWS..."
fi

# Cerca il VPC ID tramite il tag Name
VPC_ID=$(find_vpc_id "$VPC_NAME" "$AWS_REGION" "$ENDPOINT_URL" | tail -n 1) || exit 1

# Certca la Route Table ID tramite il tag Name
ROUTE_TABLE_ID=$(find_route_table_id "$ROUTE_TABLE_NAME" "$AWS_REGION" "$ENDPOINT_URL" | tail -n 1) || exit 1


# Calolcola CIDR della subnet in base alla regione
REGION_CIDR=$(aws ec2 $ENDPOINT_URL describe-vpcs --region "$AWS_REGION" --vpc-ids "$VPC_ID" --query "Vpcs[0].CidrBlock" --output text)
if [[ -z "$REGION_CIDR" || "$REGION_CIDR" == "None" ]]; then
  echo "Errore: impossibile recuperare il CIDR della VPC $VPC_NAME."
  exit 1
fi

if [[ "$COMPONENT" == "all" || "$COMPONENT" == "subnet" ]]; then

  # Controllo se la subnet esiste già
  EXISTING_SUBNET_ID=$(aws ec2 $ENDPOINT_URL describe-subnets --region "$AWS_REGION" --filters "Name=tag:Name,Values=$SUBNET_NAME" "Name=vpc-id,Values=$VPC_ID" --query "Subnets[0].SubnetId" --output text)
  if [[ -n "$EXISTING_SUBNET_ID" && "$EXISTING_SUBNET_ID" != "None" ]]; then
    echo "Subnet con nome $SUBNET_NAME già esistente con ID: $EXISTING_SUBNET_ID. Salto la creazione."
    SUBNET_CIDR=$(aws ec2 $ENDPOINT_URL describe-subnets --region "$AWS_REGION" --subnet-ids "$EXISTING_SUBNET_ID" --query "Subnets[0].CidrBlock" --output text)
    echo "CIDR della subnet esistente: $SUBNET_CIDR"
  else

    echo "Subnet con nome $SUBNET_NAME non trovata. Procedo con il calcolo del CIDR."

    # Prende tutte le sobnet create finora e calcola il prossimo CIDR disponibile
    # Per semplicità, assume che le subnet siano /24 e calcola la prossima
    # disponibile incrementando l'ultimo ottetto.
    # Esempio: se la VPC è 10.0.3.0/24 e ci sono già 3 subnet
    # (10.0.1.0/24, 10.0.2.0/24, 10.0.3.0/24), la prossima sarà
    # 10.0.4.0/24
    #!/bin/bash

    # Recupera tutti i CIDR delle subnet esistenti nella VPC
    SUBNETS_CIDRS_RAW=$(aws ec2 $ENDPOINT_URL describe-subnets --region "$AWS_REGION" --filters "Name=vpc-id,Values=$VPC_ID" --query "Subnets[].CidrBlock" --output text)

    # Converte in array
    if [[ -z "$SUBNETS_CIDRS_RAW" ]]; then
      SUBNETS_CIDRS=()
    else
      read -r -a SUBNETS_CIDRS <<< "$SUBNETS_CIDRS_RAW"
    fi

    echo "Subnet già esistenti nella VPC $VPC_NAME:"
    for CIDR in "${SUBNETS_CIDRS[@]}"; do
      echo "  - $CIDR"
    done

    # Calcola il prossimo CIDR disponibile
    # Assume che la VPC sia qualcosa come 10.0.0.0/16 e tutte le subnet /24
    IFS='.' read -r O1 O2 O3 O4 <<< "${REGION_CIDR%%/*}"
    # Per VPC /16, prendiamo O1.O2 come base
    BASE="${O1}.${O2}."
    NEXT_SUBNET=1

    while true; do
      CANDIDATE_CIDR="${BASE}${NEXT_SUBNET}.0/24"

      # Controlla se già esiste
      FOUND=false
      for EXISTING in "${SUBNETS_CIDRS[@]}"; do
        if [[ "$EXISTING" == "$CANDIDATE_CIDR" ]]; then
          FOUND=true
          break
        fi
      done

      if ! $FOUND; then
        SUBNET_CIDR="$CANDIDATE_CIDR"
        break
      fi

      ((NEXT_SUBNET++))
      if [[ $NEXT_SUBNET -gt 254 ]]; then
        echo "Errore: impossibile trovare un CIDR disponibile per la subnet."
        exit 1
      fi
    done

    echo "Calcolato CIDR per la nuova subnet $SUBNET_NAME: $SUBNET_CIDR"
  fi

  echo "Componente specificato: $COMPONENT. Eseguo il deploy del template Subnet."

  # Deploy del template
  echo "Deploy del template Subnet..."

  echo "Parametri usati:"
  echo "  Regione AWS: $AWS_REGION"
  echo "  Regione: $REGION"
  echo "  Macrozona: $MACROZONE"
  echo "  Nome Stack: $VPC_STACK_NAME"
  echo "  Nome VPC: $VPC_NAME"
  echo "  Nome Subnet: $SUBNET_NAME"
  echo "  Nome Security Group: $SECURITY_GROUP_NAME"
  echo "  Nome Route Table: $ROUTE_TABLE_NAME"

  echo "Parametri trovati:"
  echo "  VPC ID: $VPC_ID"
  echo "  Route Table ID: $ROUTE_TABLE_ID"
  echo "  CIDR Subnet: $SUBNET_CIDR"

  aws cloudformation deploy \
    --template-file "$SUBNET_TEMPLATE" \
    --stack-name "$VPC_STACK_NAME" \
    --region "$AWS_REGION" \
    --capabilities CAPABILITY_NAMED_IAM \
    --parameter-overrides VpcId="$VPC_ID" SubnetName="$SUBNET_NAME" SubnetCidr="$SUBNET_CIDR" SecurityGroupName="$SECURITY_GROUP_NAME" RouteTableId="$ROUTE_TABLE_ID" \
    $ENDPOINT_URL
  if [[ $? -ne 0 ]]; then
    echo "Errore nel deploy del template Subnet."
    aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$VPC_STACK_NAME"
    exit 1
  fi
  echo "Deploy del template Subnet completato con successo."

  # Verifica lo stato dello stack
  echo "Verifica lo stato dello stack..."
  STACK_STATUS=$(aws cloudformation $ENDPOINT_URL describe-stacks --region "$AWS_REGION" --stack-name "$VPC_STACK_NAME" --query "Stacks[0].StackStatus" --output text)
  if [[ "$STACK_STATUS" != "CREATE_COMPLETE" && "$STACK_STATUS" != "UPDATE_COMPLETE" ]]; then
    echo "Errore: lo stack $VPC_STACK_NAME non è in stato CREATE_COMPLETE o UPDATE_COMPLETE. Stato attuale: $STACK_STATUS"
    aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$VPC_STACK_NAME"
    exit 1
  fi
  echo "Lo stack $VPC_STACK_NAME è in stato $STACK_STATUS."

  # Verifica la creazione delle risorse
  echo "Verifica la creazione delle risorse..."

  echo "Verifica della Subnet..."
  SUBNET_ID=$(aws ec2 $ENDPOINT_URL describe-subnets --region "$AWS_REGION" --filters "Name=tag:Name,Values=$SUBNET_NAME" --query "Subnets[0].SubnetId" --output text)
  if [[ -z "$SUBNET_ID" || "$SUBNET_ID" == "None" ]]; then
    echo "Errore: subnet con nome $SUBNET_NAME non trovata."
    exit 1
  fi
  echo "Trovato Subnet ID: $SUBNET_ID per Subnet $SUBNET_NAME"

  echo "Verifica del Security Group..."
  SECURITY_GROUP_ID=$(aws ec2 $ENDPOINT_URL describe-security-groups --region "$AWS_REGION" --filters "Name=tag:Name,Values=$SECURITY_GROUP_NAME" --query "SecurityGroups[0].GroupId" --output text)
  if [[ -z "$SECURITY_GROUP_ID" || "$SECURITY_GROUP_ID" == "None" ]]; then
    echo "Errore: Security Group con nome $SECURITY_GROUP_NAME non trovato."
    exit 1
  fi
  echo "Trovato Security Group ID: $SECURITY_GROUP_ID per Security Group $SECURITY_GROUP_NAME"
fi

# -----------------------------
# Deploy del template Services
# -----------------------------

SERVICES_STACK_NAME="$REGION-$MACROZONE-services-stack"
SERVICES_NAME="$REGION-$MACROZONE-services"
SERVICES_HOSTNAME="$MACROZONE.mqtt-broker.${REGION}.sensor-continuum.local"
HOSTED_ZONE_NAME="$REGION.sensor-continuum.local"

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

# Creazione KeyPair per proximity services se necessario
KEYS_DIR="./keys"
SERVICES_KEY_NAME="$REGION-$MACROZONE-services-key"
SERVICES_KEY_FILE="$KEYS_DIR/$SERVICES_KEY_NAME.pem"

mkdir -p "$KEYS_DIR"

SERVICES_KEY_PAIR=$(
  { ensure_key_pair "$SERVICES_KEY_NAME" "$SERVICES_KEY_FILE" "$ENDPOINT_URL"; } | tee /dev/tty | tail -n 1
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

if [[ "$COMPONENT" == "all" || "$COMPONENT" == "services" ]]; then
  echo "Componente specificato: $COMPONENT. Eseguo il deploy del template proximity services."

  # Deploy stack CloudFormation
  echo "Deploy del template proximity services..."
  echo "Parametri usati:"
  echo "  Regione AWS: $AWS_REGION"
  echo "  Regione: $REGION"
  echo "  Macrozona: $MACROZONE"
  echo "  Nome Stack: $SERVICES_STACK_NAME"
  echo "  Tipo di istanza: $INSTANCE_TYPE"
  echo "  Subnet ID: $SUBNET_ID"
  echo "  Security Group ID: $SECURITY_GROUP_ID"
  echo "  Nome Key Name: $SERVICES_KEY_NAME"
  echo "  Nome istanza EC2: $SERVICES_NAME"
  echo "  Hosted Zone Name: $HOSTED_ZONE_NAME"
  echo "  Hostname Broker: $SERVICES_HOSTNAME"
  echo "  Environment file: $ENV_FILE"

  echo "Parametri calcolati:"
  echo "  VPC ID: $VPC_ID"
  echo "  Subnet ID: $SUBNET_ID"
  echo "  Security Group ID: $SECURITY_GROUP_ID"
  echo "  AMI ID: $IMAGE_ID"
  echo "  Hosted Zone ID: $HOSTED_ZONE_ID"

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
      KeyName="$SERVICES_KEY_NAME" \
      ServicesInstanceName="$SERVICES_NAME" \
      ServicesInstanceHostname="$SERVICES_HOSTNAME" \
      HostedZoneId="$HOSTED_ZONE_ID" \
      EnvironmentFile="$ENV_FILE" \
    $ENDPOINT_URL

  if [[ $? -ne 0 ]]; then
    echo "Errore nel deploy dello stack proximity services."
    aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$SERVICES_STACK_NAME"
    exit 1
  fi
  echo "Deploy del template proximity services completato con successo."

  # Verifica lo stato dello stack
  STACK_STATUS=$(aws cloudformation $ENDPOINT_URL describe-stacks --region "$AWS_REGION" --stack-name "$SERVICES_STACK_NAME" --query "Stacks[0].StackStatus" --output text)
  if [[ "$STACK_STATUS" != "CREATE_COMPLETE" && "$STACK_STATUS" != "UPDATE_COMPLETE" ]]; then
    echo "Errore: lo stack $SERVICES_STACK_NAME non è in stato CREATE_COMPLETE o UPDATE_COMPLETE. Stato attuale: $STACK_STATUS"
    aws $ENDPOINT_URL cloudformation describe-stack-events --stack-name "$SERVICES_STACK_NAME"
    exit 1
  fi
  echo "Lo stack $SERVICES_STACK_NAME è in stato $STACK_STATUS."

  # Recupera informazioni sull'istanza proximity services
  SERVICES_INSTANCE_ID=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --filters "Name=tag:Name,Values=$SERVICES_NAME" "Name=subnet-id,Values=$SUBNET_ID" "Name=instance-state-name,Values=pending,running,stopping,stopped" --query "Reservations[0].Instances[0].InstanceId" --output text)
  if [[ -z "$SERVICES_INSTANCE_ID" || "$SERVICES_INSTANCE_ID" == "None" ]]; then
    echo "Errore: istanza EC2 con tag Name=$SERVICES_NAME non trovata nella Subnet $SUBNET_ID."
    exit 1
  fi
  echo "Trovata istanza EC2 ID: $SERVICES_INSTANCE_ID per proximity services $SERVICES_NAME"

  # Recupera IP privato e pubblico
  SERVICES_INSTANCE_PRIVATE_IP=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --instance-ids "$SERVICES_INSTANCE_ID" --query "Reservations[0].Instances[0].PrivateIpAddress" --output text)
  SERVICES_INSTANCE_PUBLIC_IP=$(aws ec2 $ENDPOINT_URL describe-instances --region "$AWS_REGION" --instance-ids "$SERVICES_INSTANCE_ID" --query "Reservations[0].Instances[0].PublicIpAddress" --output text)

  echo "Proximity services info:"
  echo "  Private IP: $SERVICES_INSTANCE_PRIVATE_IP"
  echo "  Public IP: $SERVICES_INSTANCE_PUBLIC_IP"
  echo "  Hostname: $SERVICES_HOSTNAME"
fi

echo "Deploy completato."
