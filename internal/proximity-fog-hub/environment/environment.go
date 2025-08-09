package environment

import (
	"SensorContinuum/configs/kafka"
	"SensorContinuum/configs/mosquitto"
	"SensorContinuum/pkg/logger"
	"errors"
	"github.com/google/uuid"
	"os"
)

var Region string
var EdgeMacrozone string
var HubID string

var MosquittoProtocol string
var MosquittoBroker string
var MosquittoPort string
var FilteredDataTopic string
var HubConfigurationTopic string

var KafkaBroker string
var KafkaPort string
var ProximityDataTopic string
var ProximityConfigurationTopic string
var ProximityDataTopicPartition string
var KafkaAggregatedStatsTopic string // --- NUOVA VARIABILE ---

// --- NUOVE VARIABILI PER POSTGRES ---
var PostgresUser string
var PostgresPass string
var PostgresHost string
var PostgresPort string
var PostgresDatabase string

// --- FINE NUOVE VARIABILI ---

func SetupEnvironment() error {

	var exists bool

	Region, exists = os.LookupEnv("REGION")
	if !exists {
		return errors.New("environment variable REGION not set")
	}

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
	HubConfigurationTopic = "$share/proximity-fog-hub/configuration/hub/" + EdgeMacrozone

	KafkaBroker, exists = os.LookupEnv("KAFKA_BROKER_ADDRESS")
	if !exists {
		KafkaBroker = kafka.BROKER
	}

	KafkaPort, exists = os.LookupEnv("KAFKA_BROKER_PORT")
	if !exists {
		KafkaPort = kafka.PORT
	}

	ProximityDataTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_DATA_TOPIC")
	if !exists {
		ProximityDataTopic = kafka.PROXIMITY_FOG_HUB_DATA_TOPIC + "_" + EdgeMacrozone
	}

	ProximityConfigurationTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC")
	if !exists {
		ProximityConfigurationTopic = kafka.PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC + "_" + EdgeMacrozone
	}

	ProximityDataTopicPartition, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_TOPIC_PARTITION")
	if !exists {
		ProximityDataTopicPartition = EdgeMacrozone
	}

	// --- NUOVA LOGICA ---
	KafkaAggregatedStatsTopic, exists = os.LookupEnv("KAFKA_AGGREGATED_STATS_TOPIC")
	if !exists {
		KafkaAggregatedStatsTopic = kafka.AGGREGATED_STATS_TOPIC + "_" + EdgeMacrozone
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

	/* ----- LOGGER SETTINGS ----- */

	if err := logger.LoadLoggerFromEnv(); err != nil {
		return err
	}

	return nil

}
