#!/bin/bash

./setup_bucket.sh
if [ $? -ne 0 ]; then
  echo "Errore nella creazione del bucket."
  exit 1
fi



./deploy_region.sh region-002 --aws-region us-west-2
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della regione."
  exit 1
fi



./deploy_macrozone.sh region-002 build-0004 --aws-region us-west-2
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della macrozona."
  exit 1
fi


./deploy_zone.sh region-002 build-0004 floor-001 --aws-region us-west-2
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della zona."
  exit 1
fi

./deploy_zone.sh region-002 build-0004 floor-002 --aws-region us-west-2
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della zona."
  exit 1
fi



./deploy_macrozone.sh region-002 macrozone-0006 --aws-region us-west-2
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della macrozona."
  exit 1
fi


./deploy_zone.sh region-002 macrozone-0006 zone-001 --aws-region us-west-2
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della zona."
  exit 1
fi

./deploy_zone.sh region-002 macrozone-0006 zone-002 --aws-region us-west-2
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della zona."
  exit 1
fi