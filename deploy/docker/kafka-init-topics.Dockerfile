FROM bitnami/kafka:latest

COPY configs/kafka/init-topics.sh /init-topics.sh

ENTRYPOINT ["/init-topics.sh"]