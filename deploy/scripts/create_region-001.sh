#!/bin/bash

./setup_bucket.sh
if [ $? -ne 0 ]; then
  echo "Errore nella creazione del bucket."
  exit 1
fi



./deploy_region.sh region-001 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della regione."
  exit 1
fi



./deploy_macrozone.sh region-001 build-0001 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della macrozona."
  exit 1
fi


./deploy_zone.sh region-001 build-0001 floor-001 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della zona."
  exit 1
fi

./deploy_zone.sh region-001 build-0001 floor-002 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della zona."
  exit 1
fi



./deploy_macrozone.sh region-001 macrozone-0003 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della macrozona."
  exit 1
fi


./deploy_zone.sh region-001 macrozone-0003 zone-001 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della zona."
  exit 1
fi

./deploy_zone.sh region-001 macrozone-0003 zone-002 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della zona."
  exit 1
fi