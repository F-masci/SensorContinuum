#!/bin/bash

# ./create_bucket.sh
./create_region.sh region-001 --aws-region us-east-1
./create_macrozone.sh region-001 build-0001 --aws-region us-east-1

./create_edge_hub.sh region-001 build-0001 floor-001 --aws-region us-east-1 --instance-type t2.micro