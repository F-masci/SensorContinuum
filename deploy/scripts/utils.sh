#!/bin/bash

# Trova l'ID del VPC tramite il tag Name
find_vpc_id() {
  local vpc_name="$1"
  local region="$2"
  local endpoint="$3"
  echo "Recupero ID del VPC $vpc_name..."
  local vpc_id
  vpc_id=$(aws ec2 $endpoint describe-vpcs --region "$region" \
    --filters "Name=tag:Name,Values=$vpc_name" \
    --query "Vpcs[0].VpcId" --output text)
  if [[ -z "$vpc_id" || "$vpc_id" == "None" ]]; then
    echo "Errore: VPC con tag Name=$vpc_name non trovato."
    return 1
  fi
  echo "Trovato VPC ID: $vpc_id per VPC $vpc_name"
  echo "$vpc_id"
}

# Trova l'ID della Subnet tramite il tag Name
find_subnet_id() {
  local subnet_name="$1"
  local vpc_id="$2"
  local region="$3"
  local endpoint="$4"
  echo "Recupero ID della Subnet $subnet_name..."
  local subnet_id
  subnet_id=$(aws ec2 $endpoint describe-subnets --region "$region" \
    --filters "Name=tag:Name,Values=$subnet_name" "Name=vpc-id,Values=$vpc_id" \
    --query "Subnets[0].SubnetId" --output text)
  if [[ -z "$subnet_id" || "$subnet_id" == "None" ]]; then
    echo "Errore: Subnet con tag Name=$subnet_name non trovata nella VPC $vpc_id."
    return 1
  fi
  echo "Trovato Subnet ID: $subnet_id per Subnet $subnet_name"
  echo "$subnet_id"
}

# Trova l'ID del Security Group tramite il tag Name
find_sg_id() {
  local sg_name="$1"
  local vpc_id="$2"
  local region="$3"
  local endpoint="$4"
  echo "Recupero ID del Security Group $sg_name..."
  local sg_id
  sg_id=$(aws ec2 $endpoint describe-security-groups --region "$region" \
    --filters "Name=tag:Name,Values=$sg_name" "Name=vpc-id,Values=$vpc_id" \
    --query "SecurityGroups[0].GroupId" --output text)
  if [[ -z "$sg_id" || "$sg_id" == "None" ]]; then
    echo "Errore: Security Group con tag Name=$sg_name non trovato nella VPC $vpc_id."
    return 1
  fi
  echo "Trovato Security Group ID: $sg_id per Security Group $sg_name"
  echo "$sg_id"
}

# Trova l'ID della Route Table tramite il tag Name
find_route_table_id() {
  local rt_name="$1"
  local region="$2"
  local endpoint="$3"
  echo "Recupero ID della Route Table $rt_name..."
  local rt_id
  rt_id=$(aws ec2 $endpoint describe-route-tables --region "$region" \
    --filters "Name=tag:Name,Values=$rt_name" \
    --query "RouteTables[0].RouteTableId" --output text)
  if [[ -z "$rt_id" || "$rt_id" == "None" ]]; then
    echo "Errore: Route Table con tag Name=$rt_name non trovata."
    return 1
  fi
  echo "Trovato Route Table ID: $rt_id per Route Table $rt_name"
  echo "$rt_id"
}

# Verifica o crea la key pair e restituisce il nome
ensure_key_pair() {
  local key_name="$1"
  local key_file="$2"
  local endpoint="$3"
  echo "Verifica presenza chiave privata SSH..."
  local key
  key=$(aws ec2 $endpoint describe-key-pairs --key-names "$key_name" --query "KeyPairs[0].KeyName" --output text 2>/dev/null || true)
  if [[ -z "$key" || "$key" == "None" ]]; then
    echo "Chiave SSH $key_name non trovata in AWS. Creazione nuova chiave..."
    aws ec2 $endpoint create-key-pair \
      --key-name "$key_name" \
      --query "KeyMaterial" \
      --output text > "$key_file"
    chmod 600 "$key_file"
    echo "Chiave privata creata e salvata in $key_file"
  else
    echo "Chiave SSH $key_name già esistente in AWS."
  fi
  echo "$key_name"
}

# Trova l'AMI di Amazon Linux 2
find_amazon_linux_2_ami() {
  local region="$1"
  local endpoint="$2"
  local deploy_mode="$3"
  echo "Recupero ID dell'AMI di Amazon Linux 2..."
  local image_id
  if [[ "$deploy_mode" == "localstack" ]]; then
    image_id="ami-0c02fb55956c7d316"
    echo "Usata AMI fittizia $image_id per LocalStack."
  else
    image_id=$(aws ec2 $endpoint describe-images --region "$region" --owners amazon \
      --filters "Name=name,Values=amzn2-ami-hvm-*-x86_64-gp2" "Name=state,Values=available" \
      --query "sort_by(Images, &CreationDate)[-1].ImageId" --output text 2>/dev/null)
    if [[ -z "$image_id" || "$image_id" == "None" ]]; then
      echo "Errore: nessuna AMI di Amazon Linux 2 trovata."
      return 1
    fi
    echo "Trovata AMI di Amazon Linux 2: $image_id"
  fi
  echo "$image_id"
}