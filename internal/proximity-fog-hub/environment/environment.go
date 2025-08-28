package environment

import (
	"SensorContinuum/configs/kafka"
	"SensorContinuum/configs/mosquitto"
	"SensorContinuum/pkg/logger"
	"errors"
	"os"

	"github.com/google/uuid"
)

var EdgeMacrozone string
var HubID string

var MosquittoProtocol string
var MosquittoBroker string
var MosquittoPort string
var FilteredDataTopic string
var HubConfigurationTopic string
var HeartbeatTopic string

var KafkaBroker string
var KafkaPort string
var ProximityRealtimeDataTopic string
var ProximityConfigurationTopic string
var ProximityRealtimeDataTopicPartition string
var ProximityAggregatedStatsTopic string
var ProximityHeartbeatTopic string

// VARIABILI PER POSTGRES
var PostgresUser string
var PostgresPass string
var PostgresHost string
var PostgresPort string
var PostgresDatabase string

var HealthzServer bool = false
var HealthzServerPort string = ":"

func SetupEnvironment() error {

	var exists bool

	EdgeMacrozone, exists = os.LookupEnv("EDGE_MACROZONE")
	if !exists {
		return errors.New("environment variable EDGE_MACROZONE not set")
	}

	HubID, exists = os.LookupEnv("HUB_ID")
	if !exists {
		HubID = uuid.New().String()
	}

	MosquittoProtocol, exists = os.LookupEnv("MQTT_BROKER_PROTOCOL")
	if !exists {
		MosquittoProtocol = mosquitto.PROTOCOL
	}

	MosquittoBroker, exists = os.LookupEnv("MQTT_BROKER_ADDRESS")
	if !exists {
		MosquittoBroker = mosquitto.BROKER
	}

	MosquittoPort, exists = os.LookupEnv("MQTT_BROKER_PORT")
	if !exists {
		MosquittoPort = mosquitto.PORT
	}

	FilteredDataTopic = "$share/proximity-fog-hub/filtered-data/" + EdgeMacrozone
	HubConfigurationTopic = "configuration/hub/" + EdgeMacrozone
	HeartbeatTopic = "heartbeat/" + EdgeMacrozone

	KafkaBroker, exists = os.LookupEnv("KAFKA_BROKER_ADDRESS")
	if !exists {
		KafkaBroker = kafka.BROKER
	}

	KafkaPort, exists = os.LookupEnv("KAFKA_BROKER_PORT")
	if !exists {
		KafkaPort = kafka.PORT
	}

	ProximityRealtimeDataTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_REALTIME_DATA_TOPIC")
	if !exists {
		ProximityRealtimeDataTopic = kafka.PROXIMITY_FOG_HUB_REALTIME_DATA_TOPIC
	}

	ProximityRealtimeDataTopicPartition, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_TOPIC_PARTITION")
	if !exists {
		ProximityRealtimeDataTopicPartition = EdgeMacrozone
	}

	ProximityConfigurationTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC")
	if !exists {
		ProximityConfigurationTopic = kafka.PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC
	}

	ProximityAggregatedStatsTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_AGGREGATED_STATS_TOPIC")
	if !exists {
		ProximityAggregatedStatsTopic = kafka.PROXIMITY_FOG_HUB_AGGREGATED_STATS_TOPIC
	}

	ProximityHeartbeatTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_HEARTBEAT_TOPIC")
	if !exists {
		ProximityHeartbeatTopic = kafka.PROXIMITY_FOG_HUB_HEARTBEAT_TOPIC
	}

	PostgresUser, exists = os.LookupEnv("POSTGRES_USER")
	if !exists {
		PostgresUser = "admin"
	}

	PostgresPass, exists = os.LookupEnv("POSTGRES_PASSWORD")
	if !exists {
		PostgresPass = "adminpass"
	}

	PostgresHost, exists = os.LookupEnv("POSTGRES_HOST")
	if !exists {
		PostgresHost = "localhost" // Deve puntare al servizio del DB
	}

	PostgresPort, exists = os.LookupEnv("POSTGRES_PORT")
	if !exists {
		PostgresPort = "5432"
	}

	PostgresDatabase, exists = os.LookupEnv("POSTGRES_DATABASE")
	if !exists {
		PostgresDatabase = "sensorcontinuum"
	}

	/* ----- HEALTH CHECK SERVER SETTINGS ----- */

	HealthzServerStr, exists := os.LookupEnv("HEALTHZ_SERVER")
	if !exists {
		HealthzServer = false
	} else {
		switch HealthzServerStr {
		case "true":
			HealthzServer = true
		case "false":
			HealthzServer = false
		default:
			return errors.New("invalid value for HEALTHZ_SERVER: " + HealthzServerStr + ". Must be 'true' or 'false'")
		}
	}

	HealthzServerPort, exists = os.LookupEnv("HEALTHZ_SERVER_PORT")
	if !exists {
		HealthzServerPort = "8080"
	}

	/* ----- LOGGER SETTINGS ----- */

	if err := logger.LoadLoggerFromEnv(); err != nil {
		return err
	}

	return nil

}
