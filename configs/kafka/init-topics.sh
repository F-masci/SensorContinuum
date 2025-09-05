#!/bin/bash
set -e

echo "Creating Kafka topics if not exists..."

# aggregated-data-proximity-fog-hub
kafka-topics.sh --create --if-not-exists --topic aggregated-data-proximity-fog-hub \
  --bootstrap-server kafka-01:9092 \
  --partitions 5 --replication-factor 1

# configuration-proximity-fog-hub
kafka-topics.sh --create --if-not-exists --topic configuration-proximity-fog-hub \
  --bootstrap-server kafka-01:9092 \
  --partitions 3 --replication-factor 1

# statistics-data-proximity-fog-hub
kafka-topics.sh --create --if-not-exists --topic statistics-data-proximity-fog-hub \
  --bootstrap-server kafka-01:9092 \
  --partitions 5 --replication-factor 1

# heartbeats-proximity-fog-hub (compacted)
kafka-topics.sh --create --if-not-exists --topic heartbeats-proximity-fog-hub \
  --bootstrap-server kafka-01:9092 \
  --partitions 5 --replication-factor 1 \
  --config cleanup.policy=compact,delete
