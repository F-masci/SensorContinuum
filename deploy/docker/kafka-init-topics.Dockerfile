FROM bitnami/kafka:latest

COPY configs/kafka/init-topics.sh /init-topics.sh:ro

ENTRYPOINT ["/init-topics.sh"]