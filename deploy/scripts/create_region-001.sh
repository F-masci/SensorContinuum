#!/bin/bash

./create_bucket.sh
if [ $? -ne 0 ]; then
  echo "Errore nella creazione del bucket."
  exit 1
fi

./create_region.sh region-001 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della regione."
  exit 1
fi

./create_macrozone.sh region-001 build-0001 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della macrozona."
  exit 1
fi


./create_zone.sh region-001 build-0001 floor-001 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della zona."
  exit 1
fi