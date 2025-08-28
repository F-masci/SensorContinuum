package environment

import (
	"SensorContinuum/configs/kafka"
	"SensorContinuum/pkg/logger"
	"errors"
	"os"

	"github.com/google/uuid"
)

var HubID string

var KafkaBroker string
var KafkaPort string
var KafkaAggregatedStatsTopic string
var ProximityDataTopic string
var ProximityConfigurationTopic string
var ProximityHeartbeatTopic string
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

// Configurazioni batch
var SensorDataBatchSize int = 10    // Dimensione del batch per i dati dei sensori
var SensorDataBatchTimeout int = 10 // Timeout in secondi per il batch dei dati dei sensori

var HealthzServer bool = false
var HealthzServerPort string = ":"

func SetupEnvironment() error {

	var exists bool

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

	ProximityDataTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_REALTIME_DATA_TOPIC")
	if !exists {
		ProximityDataTopic = kafka.PROXIMITY_FOG_HUB_REALTIME_DATA_TOPIC
	}

	ProximityConfigurationTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC")
	if !exists {
		ProximityConfigurationTopic = kafka.PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC
	}

	ProximityHeartbeatTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_HEARTBEAT_TOPIC")
	if !exists {
		ProximityHeartbeatTopic = kafka.PROXIMITY_FOG_HUB_HEARTBEAT_TOPIC
	}

	IntermediateDataTopic, exists = os.LookupEnv("KAFKA_INTERMEDIATE_FOG_HUB_TOPIC")
	if !exists {
		IntermediateDataTopic = kafka.INTERMEDIATE_FOG_HUB_TOPIC
	}

	KafkaAggregatedStatsTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_AGGREGATED_STATS_TOPIC")
	if !exists {
		KafkaAggregatedStatsTopic = kafka.PROXIMITY_FOG_HUB_AGGREGATED_STATS_TOPIC
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
