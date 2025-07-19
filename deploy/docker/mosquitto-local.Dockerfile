FROM eclipse-mosquitto:latest

COPY configs/mosquitto/mosquitto-local.conf /mosquitto/config/mosquitto.conf