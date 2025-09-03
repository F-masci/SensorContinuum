#!/bin/bash

LOG_FILE="/var/log/docker-install.log"
exec > >(tee -a "$LOG_FILE") 2>&1

echo "Inizio installazione Docker e Docker Compose..."

# Aggiorna i pacchetti di sistema
sudo yum update -y

# Installa Docker tramite amazon-linux-extras
sudo amazon-linux-extras enable docker
sudo amazon-linux-extras install docker -y

# Avvia il servizio Docker
sudo systemctl start docker
sudo systemctl enable docker

# Aggiungi l'utente ec2-user al gruppo docker per evitare sudo
sudo usermod -aG docker ec2-user

# Installa Docker Compose
DOCKER_COMPOSE_VERSION="v2.39.2"
sudo curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verifica le installazioni
docker --version
docker-compose --version

echo "Docker e Docker Compose installati correttamente!"