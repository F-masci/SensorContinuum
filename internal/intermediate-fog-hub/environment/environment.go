package environment

import (
	"SensorContinuum/configs/kafka"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// OperationMode specifica la modalità di funzionamento del servizio.
var OperationMode types.OperationModeType

// ServiceMode specifica il tipo di servizio in esecuzione.
var ServiceMode types.Service

var HubID string

// KafkaBroker specifica l'indirizzo del broker Kafka.
var KafkaBroker string

// KafkaPort specifica la porta del broker Kafka.
var KafkaPort string

// ProximityDataTopic specifica il topic Kafka per i dati in tempo reale dai sensori.
var ProximityDataTopic string

// AggregatedStatsTopic specifica il topic Kafka per i dati statistici aggregati.
var AggregatedStatsTopic string

// ProximityConfigurationTopic specifica il topic Kafka per i messaggi di configurazione.
var ProximityConfigurationTopic string

// ProximityHeartbeatTopic specifica il topic Kafka per i messaggi di heartbeat.
var ProximityHeartbeatTopic string

// KafkaCommitTimeout specifica il timeout per il commit degli offset Kafka.
var KafkaCommitTimeout int = 5

// Queste impostazioni sono utilizzate per la connessione ai databases PostgreSQL.

/* ------ POSTGRESQL DATABASES ------ */
/*				Region DB			  */
/* ---------------------------------- */

// PostgresRegionUser specifica l'utente per il database PostgreSQL dei metadati della regione.
var PostgresRegionUser string

// PostgresRegionPass specifica la password per il database PostgreSQL dei metadati della regione.
var PostgresRegionPass string

// PostgresRegionHost specifica l'host per il database PostgreSQL dei metadati della regione.
var PostgresRegionHost string

// PostgresRegionPort specifica la porta per il database PostgreSQL dei metadati della regione.
var PostgresRegionPort string

// PostgresRegionDatabase specifica il nome del database PostgreSQL dei metadati della regione.
var PostgresRegionDatabase string

/* ------ POSTGRESQL DATABASES ------ */
/*				Cloud DB			  */
/* ---------------------------------- */

// PostgresCloudUser specifica l'utente per il database PostgreSQL del cloud.
var PostgresCloudUser string

// PostgresCloudPass specifica la password per il database PostgreSQL del cloud.
var PostgresCloudPass string

// PostgresCloudHost specifica l'host per il database PostgreSQL del cloud.
var PostgresCloudHost string

// PostgresCloudPort specifica la porta per il database PostgreSQL del cloud.
var PostgresCloudPort string

// PostgresCloudDatabase specifica il nome del database PostgreSQL del cloud.
var PostgresCloudDatabase string

/* ------ POSTGRESQL DATABASES ------ */
/*				Sensor DB			  */
/* ---------------------------------- */

// PostgresSensorUser specifica l'utente per il database PostgreSQL dei dati dei sensori della regione.
var PostgresSensorUser string

// PostgresSensorPass specifica la password per il database PostgreSQL dei dati dei sensori della regione.
var PostgresSensorPass string

// PostgresSensorHost specifica l'host per il database PostgreSQL dei dati dei sensori della regione.
var PostgresSensorHost string

// PostgresSensorPort specifica la porta per il database PostgreSQL dei dati dei sensori della regione.
var PostgresSensorPort string

// PostgresSensorDatabase specifica il nome del database PostgreSQL dei dati dei sensori della regione.
var PostgresSensorDatabase string

// Configurazioni batch per l'invio dei dati a Kafka

// SensorDataBatchSize specifica la dimensione del batch per i dati dei sensori.
var SensorDataBatchSize int = 100

// SensorDataBatchTimeout specifica il timeout per il batch dei dati dei sensori.
var SensorDataBatchTimeout int = 15

// AggregatedDataBatchSize specifica la dimensione del batch per i dati aggregati.
var AggregatedDataBatchSize int = 100

// AggregatedDataBatchTimeout specifica il timeout per il batch dei dati aggregati.
var AggregatedDataBatchTimeout int = 15

const (
	// KafkaGroupId specifica il group ID per i consumer Kafka.
	// Poiché il fog hub gestisce una singola regione, tutti i servizi usanono lo stesso group ID.
	// In questo modo, ogni messaggio viene elaborato da un solo servizio e non duplicato.
	// Kafka gestisce il bilanciamento del carico tra i nodi del servizio.
	KafkaGroupId = "intermediate-fog-hub"

	// AggregationInterval specifica l'intervallo di tempo per l'aggregazione dei dati.
	AggregationInterval = 30 * time.Minute
	// AggregationStartingOffset è il tempo in meno per costruire il primo intervallo di aggregazione,
	// in modo da includere eventuali dati ricevuti prima dell'avvio del servizio.
	AggregationStartingOffset = -24 * time.Hour
	// AggregationFetchOffset è il tempo extra per recuperare i dati dai sensori,
	// in modo da includere eventuali ritardi nella ricezione dei messaggi.
	// Specifica l'offest negativo di tempo rispetto all'istante corrente
	// per recuperare i dati aggregati.
	AggregationFetchOffset = -20 * time.Minute
)

var HealthzServer bool = false
var HealthzServerPort string = ":"

func SetupEnvironment() error {

	var exists bool

	/* ----- OPERATION MODE ----- */

	var OperationModeStr string
	OperationModeStr, exists = os.LookupEnv("OPERATION_MODE")
	if !exists {
		OperationMode = types.OperationModeLoop
	} else {
		switch OperationModeStr {
		case string(types.OperationModeLoop):
			OperationMode = types.OperationModeLoop
		case string(types.OperationModeOnce):
			OperationMode = types.OperationModeOnce
		default:
			return errors.New("invalid value for OPERATION_MODE: " + OperationModeStr + ". Valid values are 'loop' or 'once'.")
		}
	}

	/* ----- SERVICE MODE ----- */

	ServiceModeStr, exists := os.LookupEnv("SERVICE_MODE")
	if !exists {
		ServiceMode = types.IntermediateHubService
	} else {
		switch ServiceModeStr {
		case string(types.IntermediateHubService):
			ServiceMode = types.IntermediateHubService
		case string(types.IntermediateHubAggregatorService):
			ServiceMode = types.IntermediateHubAggregatorService
		case string(types.IntermediateHubConfigurationService):
			ServiceMode = types.IntermediateHubConfigurationService
		case string(types.IntermediateHubHeartbeatService):
			ServiceMode = types.IntermediateHubHeartbeatService
		default:
			return errors.New("invalid value for SERVICE_MODE: " + ServiceModeStr + ". Valid values are 'intermediate_hub_service', 'intermediate_hub_aggregator_service', 'intermediate_hub_configuration_service' or 'intermediate_hub_heartbeat_service'.")
		}
	}

	/* ----- ENVIRONMENT SETTINGS ----- */

	HubID, exists = os.LookupEnv("HUB_ID")
	if !exists {
		HubID = uuid.New().String()
	}

	/* ----- KAFKA BROKER SETTINGS ----- */

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

	AggregatedStatsTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_AGGREGATED_STATS_TOPIC")
	if !exists {
		AggregatedStatsTopic = kafka.PROXIMITY_FOG_HUB_AGGREGATED_STATS_TOPIC
	}

	ProximityConfigurationTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC")
	if !exists {
		ProximityConfigurationTopic = kafka.PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC
	}

	ProximityHeartbeatTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_HEARTBEAT_TOPIC")
	if !exists {
		ProximityHeartbeatTopic = kafka.PROXIMITY_FOG_HUB_HEARTBEAT_TOPIC
	}

	var KafkaCommitTimeoutStr string
	KafkaCommitTimeoutStr, exists = os.LookupEnv("KAFKA_COMMIT_TIMEOUT")
	if exists {
		var err error
		KafkaCommitTimeout, err = strconv.Atoi(KafkaCommitTimeoutStr)
		if err != nil || KafkaCommitTimeout <= 0 {
			return errors.New("invalid value for KAFKA_COMMIT_TIMEOUT: " + KafkaCommitTimeoutStr + ". Must be a positive integer representing seconds.")
		}
	}

	/* ----- POSTGRESQL DATABASES SETTINGS ----- */
	/* 				  Region DB			  	 	 */
	/* ----------------------------------------- */

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

	/* ----- POSTGRESQL DATABASES SETTINGS ----- */
	/* 				  Cloud DB			  	 	 */
	/* ----------------------------------------- */

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

	/* ----- POSTGRESQL DATABASES SETTINGS ----- */
	/* 				  Sensor DB			  	 	 */
	/* ----------------------------------------- */

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

	SensorDataBatchSizeStr, exists := os.LookupEnv("SENSOR_DATA_BATCH_SIZE")
	if exists {
		var err error
		SensorDataBatchSize, err = strconv.Atoi(SensorDataBatchSizeStr)
		if err != nil || SensorDataBatchSize <= 0 {
			return errors.New("invalid value for SENSOR_DATA_BATCH_SIZE: " + SensorDataBatchSizeStr + ". Must be a positive integer.")
		}
	}
	SensorDataBatchTimeoutStr, exists := os.LookupEnv("SENSOR_DATA_BATCH_TIMEOUT")
	if exists {
		var err error
		SensorDataBatchTimeout, err = strconv.Atoi(SensorDataBatchTimeoutStr)
		if err != nil || SensorDataBatchTimeout <= 0 {
			return errors.New("invalid value for SENSOR_DATA_BATCH_TIMEOUT: " + SensorDataBatchTimeoutStr + ". Must be a positive integer.")
		}
	}

	AggregatedDataBatchSizeStr, exists := os.LookupEnv("AGGREGATED_DATA_BATCH_SIZE")
	if exists {
		var err error
		AggregatedDataBatchSize, err = strconv.Atoi(AggregatedDataBatchSizeStr)
		if err != nil || AggregatedDataBatchSize <= 0 {
			return errors.New("invalid value for AGGREGATED_DATA_BATCH_SIZE: " + AggregatedDataBatchSizeStr + ". Must be a positive integer.")
		}
	}
	AggregatedDataBatchTimeoutStr, exists := os.LookupEnv("AGGREGATED_DATA_BATCH_TIMEOUT")
	if exists {
		var err error
		AggregatedDataBatchTimeout, err = strconv.Atoi(AggregatedDataBatchTimeoutStr)
		if err != nil || AggregatedDataBatchTimeout <= 0 {
			return errors.New("invalid value for AGGREGATED_DATA_BATCH_TIMEOUT: " + AggregatedDataBatchTimeoutStr + ". Must be a positive integer.")
		}
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
