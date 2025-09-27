#!/bin/bash

set -e

# Lista regioni
./deploy_lambda.sh region-list-stack region regionList "/region/list"

# Ricerca region per nome
./deploy_lambda.sh region-search-name-stack region regionSearchName "/region/search/name/{name}"

# Dati aggregati per regione
./deploy_lambda.sh region-data-aggregated-stack region regionDataAggregated "/region/data/aggregated/{region}"


# Lista macrozone
./deploy_lambda.sh macrozone-list-stack macrozone macrozoneList "/macrozone/list/{region}"

# Ricerca macrozone per nome
./deploy_lambda.sh macrozone-search-name-stack macrozone macrozoneSearchName "/macrozone/search/name/{region}/{name}"

# Dati aggregati per macrozone
./deploy_lambda.sh macrozone-data-aggregated-name-stack macrozone macrozoneDataAggregatedName "/macrozone/data/aggregated/{region}/{macrozone}"

# Dati aggregati per location
./deploy_lambda.sh macrozone-data-aggregated-location-stack macrozone macrozoneDataAggregatedLocation "/macrozone/data/aggregated/location"

# Trend macrozone
./deploy_lambda.sh macrozone-data-trend-stack macrozone macrozoneDataTrend "/macrozone/data/trend/{region}"

# Variazione macrozone
./deploy_lambda.sh macrozone-data-variation-stack macrozone macrozoneDataVariation "/macrozone/data/variation/{region}"

# Correlazione variazione macrozone
./deploy_lambda.sh macrozone-data-variation-correlation-stack macrozone macrozoneDataVariationCorrelation "/macrozone/data/variation/correlation/{region}"



# Lista zone
./deploy_lambda.sh zone-list-stack zone zoneList "/zone/list/{region}/{macrozone}"

# Ricerca zona per nome
./deploy_lambda.sh zone-search-name-stack zone zoneSearchName "/zone/search/name/{region}/{macrozone}/{name}"

# Dati sensori raw per zona
./deploy_lambda.sh zone-sensor-data-raw-stack zone zoneSensorDataRaw "/zone/sensor/data/raw/{region}/{macrozone}/{zone}/{sensor}"

# Dati aggregati zona
./deploy_lambda.sh zone-data-aggregated-stack zone zoneDataAggregated "/zone/data/aggregated/{region}/{macrozone}/{zone}"
