#!/bin/bash

# ==================================
# FASE CLOUD
# ==================================

# 1. SETUP INIZIALE: Preparazione degli asset su AWS S3
./setup_bucket.sh
if [ $? -ne 0 ]; then
  echo "Errore nella creazione del bucket S3."
  exit 1
fi

# 2. CONFIGURAZIONE DNS PUBBLICO (CLOUD)
./setup_dns.sh
if [ $? -ne 0 ]; then
  echo "Errore nella configurazione DNS pubblica."
  exit 1
fi

# 3. SETUP LAMBDA (CLOUD)
# Rete VPC per Lambda, API Gateway, Certificati SSL.
# ./setup_lambda.sh
if [ $? -ne 0 ]; then
  echo "Errore nel deploy API e Lambda."
  exit 1
fi

# 4. DEPLOY DATABASE CLOUD (CLOUD)
# Cluster Aurora PostgreSQL per i metadati globali.
./setup_cloud_db.sh
if [ $? -ne 0 ]; then
  echo "Errore nel deploy del Database Cloud dei Metadati."
  exit 1
fi

# 5. SETUP SITO WEB (CLOUD)
# Configura l'hosting del frontend. (Il deploy del codice va fatto da console web).
./setup_site.sh
if [ $? -ne 0 ]; then
  echo "Errore nel setup del sito web."
  exit 1
fi
