kafka-topics.sh --delete --topic aggregated-data-proximity-fog-hub --bootstrap-server kafka:9092
kafka-topics.sh --delete --topic configuration-proximity-fog-hub --bootstrap-server kafka:9092
kafka-topics.sh --delete --topic statistics-data-proximity-fog-hub --bootstrap-server kafka:9092
kafka-topics.sh --delete --topic heartbeats-proximity-fog-hub --bootstrap-server kafka:9092

# aggregated-data-proximity-fog-hub
kafka-topics.sh --create --topic aggregated-data-proximity-fog-hub \
  --bootstrap-server kafka:9092 \
  --partitions 5 --replication-factor 1

# configuration-proximity-fog-hub
kafka-topics.sh --create --topic configuration-proximity-fog-hub \
  --bootstrap-server kafka:9092 \
  --partitions 3 --replication-factor 1

# statistics-data-proximity-fog-hub
kafka-topics.sh --create --topic statistics-data-proximity-fog-hub \
  --bootstrap-server kafka:9092 \
  --partitions 5 --replication-factor 1

# heartbeats-proximity-fog-hub (compacted)
kafka-topics.sh --create --topic heartbeats-proximity-fog-hub \
  --bootstrap-server kafka:9092 \
  --partitions 5 --replication-factor 1 \
  --config cleanup.policy=compact,delete
