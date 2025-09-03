#!/bin/bash

LOG_FILE="/var/log/user-data.log"
exec > >(tee -a $LOG_FILE) 2>&1

echo "UserData iniziato $(date)"

# Aggiornamento sistema
yum update -y

# Installazione Docker
amazon-linux-extras install docker -y
service docker start
usermod -a -G docker ec2-user

# Installazione Docker Compose
DOCKER_COMPOSE=/usr/local/bin/docker-compose
if [ ! -f "$DOCKER_COMPOSE" ]; then
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o "$DOCKER_COMPOSE"
    chmod +x "$DOCKER_COMPOSE"
fi

# Verifica AWS CLI
if ! command -v aws &> /dev/null; then
    echo "Installazione AWS CLI v2..."
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "/tmp/awscliv2.zip"
    unzip /tmp/awscliv2.zip -d /tmp
    /tmp/aws/install
fi

# Parametri passati da CloudFormation
BUCKET="${S3Bucket}"
PREFIX="${S3Prefix}"
SCRIPTS="${Scripts}"

echo "Scarico ed eseguo script da s3://$BUCKET/$PREFIX"
for script in $SCRIPTS; do
    echo "Scarico $script..."
    aws s3 cp "s3://$BUCKET/$PREFIX$script" "/tmp/$script"
    chmod +x "/tmp/$script"
    echo "Eseguo $script..."
    /tmp/$script
done

echo "UserData completato $(date)"
