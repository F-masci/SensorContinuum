package environment

import (
	"SensorContinuum/configs/kafka"
	"errors"
	"github.com/google/uuid"
	"os"
)

var BuildingID string
var HubID string

var KafkaBroker string
var KafkaPort string
var ProximityDataTopic string
var ProximityConfigurationTopic string
var IntermediateDataTopic string

// Variabili per il DB Region
var PostgresRegionUser string
var PostgresRegionPass string
var PostgresRegionHost string
var PostgresRegionPort string
var PostgresRegionDatabase string

// Variabili per il DB Cloud
var PostgresCloudUser string
var PostgresCloudPass string
var PostgresCloudHost string
var PostgresCloudPort string
var PostgresCloudDatabase string

// Variabili per il DB Sensor
var PostgresSensorUser string
var PostgresSensorPass string
var PostgresSensorHost string
var PostgresSensorPort string
var PostgresSensorDatabase string

func SetupEnvironment() error {

	var exists bool

	BuildingID, exists = os.LookupEnv("BUILDING_ID")
	if !exists {
		return errors.New("environment variable BUILDING_ID not set")
	}

	HubID, exists = os.LookupEnv("HUB_ID")
	if !exists {
		HubID = uuid.New().String()
	}

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
		ProximityDataTopic = kafka.PROXIMITY_FOG_HUB_DATA_TOPIC + "_" + BuildingID
	}

	ProximityConfigurationTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC")
	if !exists {
		ProximityConfigurationTopic = kafka.PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC + "_" + BuildingID
	}

	IntermediateDataTopic, exists = os.LookupEnv("KAFKA_INTERMEDIATE_FOG_HUB_TOPIC")
	if !exists {
		IntermediateDataTopic = kafka.INTERMEDIATE_FOG_HUB_TOPIC
	}

	// Inizializzazione variabili DB Region
	PostgresRegionUser, exists = os.LookupEnv("POSTGRES_REGION_USER")
	if !exists {
		PostgresRegionUser = "admin"
	}
	PostgresRegionPass, exists = os.LookupEnv("POSTGRES_REGION_PASSWORD")
	if !exists {
		PostgresRegionPass = "adminpass"
	}
	PostgresRegionHost, exists = os.LookupEnv("POSTGRES_REGION_HOST")
	if !exists {
		PostgresRegionHost = "localhost"
	}
	PostgresRegionPort, exists = os.LookupEnv("POSTGRES_REGION_PORT")
	if !exists {
		PostgresRegionPort = "5434"
	}
	PostgresRegionDatabase, exists = os.LookupEnv("POSTGRES_REGION_DATABASE")
	if !exists {
		PostgresRegionDatabase = "sensorcontinuum"
	}

	// Inizializzazione variabili DB Cloud
	PostgresCloudUser, exists = os.LookupEnv("POSTGRES_CLOUD_USER")
	if !exists {
		PostgresCloudUser = "admin"
	}
	PostgresCloudPass, exists = os.LookupEnv("POSTGRES_CLOUD_PASSWORD")
	if !exists {
		PostgresCloudPass = "adminpass"
	}
	PostgresCloudHost, exists = os.LookupEnv("POSTGRES_CLOUD_HOST")
	if !exists {
		PostgresCloudHost = "localhost"
	}
	PostgresCloudPort, exists = os.LookupEnv("POSTGRES_CLOUD_PORT")
	if !exists {
		PostgresCloudPort = "5433"
	}
	PostgresCloudDatabase, exists = os.LookupEnv("POSTGRES_CLOUD_DATABASE")
	if !exists {
		PostgresCloudDatabase = "sensorcontinuum"
	}

	// Inizializzazione variabili DB Sensor
	PostgresSensorUser, exists = os.LookupEnv("POSTGRES_SENSOR_USER")
	if !exists {
		PostgresSensorUser = "admin"
	}
	PostgresSensorPass, exists = os.LookupEnv("POSTGRES_SENSOR_PASSWORD")
	if !exists {
		PostgresSensorPass = "adminpass"
	}
	PostgresSensorHost, exists = os.LookupEnv("POSTGRES_SENSOR_HOST")
	if !exists {
		PostgresSensorHost = "localhost"
	}
	PostgresSensorPort, exists = os.LookupEnv("POSTGRES_SENSOR_PORT")
	if !exists {
		PostgresSensorPort = "5432"
	}
	PostgresSensorDatabase, exists = os.LookupEnv("POSTGRES_SENSOR_DATABASE")
	if !exists {
		PostgresSensorDatabase = "sensorcontinuum"
	}

	return nil

}
