package environment

import (
	"SensorContinuum/configs/kafka"
	"SensorContinuum/configs/mosquitto"
	"SensorContinuum/configs/timeouts"
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

var EdgeMacrozone string
var HubID string

// Queste impostazioni sono utilizzate per la connessione tra:
// - EdgeHub -> ProximityFogHub

// MqttProtocol specifica il protocollo di connessione al broker MQTT (es. "tcp", "ws").
var MqttProtocol string

// MqttBroker specifica l'indirizzo del broker MQTT.
var MqttBroker string

// MqttPort specifica la porta del broker MQTT.
var MqttPort string

// FilteredDataTopic è il topic MQTT su cui il Proximity Fog Hub riceve i dati filtrati dall'Edge Hub.
var FilteredDataTopic string

// HubConfigurationTopic è il topic MQTT su cui il Proximity Fog Hub riceve i messaggi di configurazione.
var HubConfigurationTopic string

// HeartbeatTopic è il topic MQTT su cui il Proximity Fog Hub riceve i messaggi di heartbeat.
var HeartbeatTopic string

// Queste impostazioni controllano il comportamento della riconnessione al broker MQTT.

// MqttMaxReconnectionInterval specifica l'intervallo massimo tra i tentativi di riconnessione in secondi.
var MqttMaxReconnectionInterval int = 10

// MqttMaxReconnectionTimeout specifica il timeout massimo per ogni tentativo di riconnessione in secondi.
var MqttMaxReconnectionTimeout int = 10

// MqttMaxReconnectionAttempts specifica il numero massimo di tentativi di riconnessione.
var MqttMaxReconnectionAttempts int = 10

// MqttMaxSubscriptionTimeout specifica il timeout per la sottoscrizione ai topic in secondi.
var MqttMaxSubscriptionTimeout int = 5

// MqttMessagePublishTimeout specifica il timeout per la pubblicazione dei messaggi in secondi.
var MqttMessagePublishTimeout int = 5

// MqttMessagePublishAttempts specifica il numero di tentativi di pubblicazione dei messaggi.
var MqttMessagePublishAttempts int = 3

// MqttMessageCleaningTimeout specifica il timeout per la pulizia dei messaggi in secondi.
var MqttMessageCleaningTimeout int = MqttMessagePublishTimeout

// Queste impostazioni sono utilizzate per la connessione tra:
// - ProximityFogHub -> IntermediateFogHub

// KafkaBroker specifica l'indirizzo del broker Kafka.
var KafkaBroker string

// KafkaPort specifica la porta del broker Kafka.
var KafkaPort string

// ProximityRealtimeDataTopic è il topic Kafka su cui il Proximity Fog Hub invia i dati in tempo reale all'Intermediate Fog Hub.
var ProximityRealtimeDataTopic string

// ProximityConfigurationTopic è il topic Kafka su cui il Proximity Fog Hub invia i messaggi di configurazione all'Intermediate Fog Hub.
var ProximityConfigurationTopic string

// ProximityAggregatedStatsTopic è il topic Kafka su cui il Proximity Fog Hub invia le statistiche aggregate all'Intermediate Fog Hub.
var ProximityAggregatedStatsTopic string

// ProximityHeartbeatTopic è il topic Kafka su cui il Proximity Fog Hub invia i messaggi di heartbeat all'Intermediate Fog Hub.
var ProximityHeartbeatTopic string

// KafkaPublishTimeout specifica il timeout per la pubblicazione dei messaggi su Kafka in secondi.
var KafkaPublishTimeout int = 5

// Queste impostazioni sono utilizzate per la connessione al database PostgreSQL locale.
// Il database viene utilizzato per la memorizzazione temporanea dei dati e delle statistiche.

// PostgresUser specifica l'utente per la connessione al database PostgreSQL.
var PostgresUser string

// PostgresPass specifica la password per la connessione al database PostgreSQL.
var PostgresPass string

// PostgresHost specifica l'indirizzo del server PostgreSQL.
var PostgresHost string

// PostgresPort specifica la porta del server PostgreSQL.
var PostgresPort string

// PostgresDatabase specifica il nome del database PostgreSQL.
var PostgresDatabase string

const (
	// AggregationInterval specifica l'intervallo di tempo per l'aggregazione dei dati.
	AggregationInterval = 15 * time.Minute
	// AggregationStartingOffset è il tempo in meno per costruire il primo intervallo di aggregazione,
	// in modo da includere eventuali dati ricevuti prima dell'avvio del servizio.
	AggregationStartingOffset = -24 * time.Hour
	// AggregationFetchOffset è il tempo extra per recuperare i dati dai sensori,
	// in modo da includere eventuali ritardi nella ricezione dei messaggi.
	// Specifica l'offest negativo di tempo rispetto all'istante corrente
	// per recuperare i dati aggregati.
	AggregationFetchOffset = -10 * time.Minute
	// AggregationLockId specifica l'ID del lock per l'aggregazione.
	// Serve per evitare che più istanze del servizio eseguano l'aggregazione contemporaneamente.
	AggregationLockId = 472

	// OutboxPollInterval definisce ogni quanto il dispatcher controlla la tabella outbox.
	OutboxPollInterval = 2 * time.Minute
	// OutboxBatchSize definisce quanti messaggi il dispatcher tenta di inviare in ogni ciclo.
	OutboxBatchSize = 50
	// OutboxMaxAttempts definisce il numero massimo di tentativi di invio per ogni batch di messaggi.
	OutboxMaxAttempts = 3
	// DispatcherLockId specifica l'ID del lock per il dispatcher.
	// Serve per evitare che più istanze del servizio eseguano il dispatching contemporaneamente.
	DispatcherLockId = 853

	// CleanerInterval definisce ogni quanto il cleaner si attiva per pulire la tabella outbox.
	CleanerInterval = 5 * time.Minute
	// SentMessageMaxAge definisce l'età minima che un messaggio 'sent' deve avere prima di essere eliminato.
	// Questo fornisce una finestra di sicurezza per il debug o l'auditing, evitando di cancellare
	// messaggi che sono stati appena inviati.
	SentMessageMaxAge = 12 * time.Hour

	// HeartbeatInterval specifica l'intervallo di tempo tra i messaggi di heartbeat inviati all'Intermediate Fog Hub.
	HeartbeatInterval = timeouts.HeartbeatInterval
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
		ServiceMode = types.ProximityHubService
	} else {
		switch ServiceModeStr {
		case string(types.ProximityHubService):
			ServiceMode = types.ProximityHubService
		case string(types.ProximityHubLocalCacheService):
			ServiceMode = types.ProximityHubLocalCacheService
		case string(types.ProximityHubConfigurationService):
			ServiceMode = types.ProximityHubConfigurationService
		case string(types.ProximityHubHeartbeatService):
			ServiceMode = types.ProximityHubHeartbeatService
		case string(types.ProximityHubAggregatorService):
			ServiceMode = types.ProximityHubAggregatorService
		case string(types.ProximityHubDispatcherService):
			ServiceMode = types.ProximityHubDispatcherService
		case string(types.ProximityHubCleanerService):
			ServiceMode = types.ProximityHubCleanerService
		default:
			return errors.New("invalid value for SERVICE_MODE: " + ServiceModeStr + ". Valid values are 'proximity_hub', 'proximity_hub_local_cache', 'proximity_hub_configuration', 'proximity_hub_heartbeat', 'proximity_hub_aggregator', 'proximity_hub_dispatcher', 'proximity_hub_cleaner'")
		}
	}

	/* ----- ENVIRONMENT SETTINGS ----- */

	EdgeMacrozone, exists = os.LookupEnv("EDGE_MACROZONE")
	if !exists {
		return errors.New("environment variable EDGE_MACROZONE not set")
	}

	HubID, exists = os.LookupEnv("HUB_ID")
	if !exists {
		HubID = uuid.New().String()
	}

	/* ----- MQTT BROKER SETTINGS ----- */

	MqttProtocol, exists = os.LookupEnv("MQTT_BROKER_PROTOCOL")
	if !exists {
		MqttProtocol = mosquitto.PROTOCOL
	}

	MqttBroker, exists = os.LookupEnv("MQTT_BROKER_ADDRESS")
	if !exists {
		MqttBroker = mosquitto.BROKER
	}

	MqttPort, exists = os.LookupEnv("MQTT_BROKER_PORT")
	if !exists {
		MqttPort = mosquitto.PORT
	}

	FilteredDataTopic = "$share/proximity-fog-hub_" + EdgeMacrozone + "/filtered-data/" + EdgeMacrozone
	HubConfigurationTopic = "configuration/hub/" + EdgeMacrozone
	HeartbeatTopic = "heartbeat/" + EdgeMacrozone

	var MqttMaxReconnectionIntervalStr string
	MqttMaxReconnectionIntervalStr, exists = os.LookupEnv("MQTT_MAX_RECONNECTION_INTERVAL")
	if exists {
		var err error
		MqttMaxReconnectionInterval, err = strconv.Atoi(MqttMaxReconnectionIntervalStr)
		if err != nil || MqttMaxReconnectionInterval <= 0 {
			return errors.New("invalid value for MQTT_MAX_RECONNECTION_INTERVAL: " + MqttMaxReconnectionIntervalStr + ". Must be a positive integer")
		}
	}

	var MqttMaxReconnectionTimeoutStr string
	MqttMaxReconnectionTimeoutStr, exists = os.LookupEnv("MQTT_MAX_RECONNECTION_TIMEOUT")
	if exists {
		var err error
		MqttMaxReconnectionTimeout, err = strconv.Atoi(MqttMaxReconnectionTimeoutStr)
		if err != nil || MqttMaxReconnectionTimeout <= 0 {
			return errors.New("invalid value for MQTT_MAX_RECONNECTION_TIMEOUT: " + MqttMaxReconnectionTimeoutStr + ". Must be a positive integer")
		}
	}

	var MqttMaxReconnectionAttemptsStr string
	MqttMaxReconnectionAttemptsStr, exists = os.LookupEnv("MQTT_MAX_RECONNECTION_ATTEMPTS")
	if exists {
		var err error
		MqttMaxReconnectionAttempts, err = strconv.Atoi(MqttMaxReconnectionAttemptsStr)
		if err != nil || MqttMaxReconnectionAttempts <= 0 {
			return errors.New("invalid value for MQTT_MAX_RECONNECTION_ATTEMPTS: " + MqttMaxReconnectionAttemptsStr + ". Must be a positive integer")
		}
	}

	var MqttMaxSubscriptionTimeoutStr string
	MqttMaxSubscriptionTimeoutStr, exists = os.LookupEnv("MQTT_MAX_SUBSCRIPTION_TIMEOUT")
	if exists {
		var err error
		MqttMaxSubscriptionTimeout, err = strconv.Atoi(MqttMaxSubscriptionTimeoutStr)
		if err != nil || MqttMaxSubscriptionTimeout <= 0 {
			return errors.New("invalid value for MQTT_MAX_SUBSCRIPTION_TIMEOUT: " + MqttMaxSubscriptionTimeoutStr + ". Must be a positive integer")
		}
	}

	var MqttMessagePublishTimeoutStr string
	MqttMessagePublishTimeoutStr, exists = os.LookupEnv("MQTT_MESSAGE_PUBLISH_TIMEOUT")
	if exists {
		var err error
		MqttMessagePublishTimeout, err = strconv.Atoi(MqttMessagePublishTimeoutStr)
		if err != nil || MqttMessagePublishTimeout <= 0 {
			return errors.New("invalid value for MQTT_MESSAGE_PUBLISH_TIMEOUT: " + MqttMessagePublishTimeoutStr + ". Must be a positive integer")
		}
	}

	var MqttMessagePublishAttemptsStr string
	MqttMessagePublishAttemptsStr, exists = os.LookupEnv("MQTT_MESSAGE_PUBLISH_ATTEMPTS")
	if exists {
		var err error
		MqttMessagePublishAttempts, err = strconv.Atoi(MqttMessagePublishAttemptsStr)
		if err != nil || MqttMessagePublishAttempts <= 0 {
			return errors.New("invalid value for MQTT_MESSAGE_PUBLISH_ATTEMPTS: " + MqttMessagePublishAttemptsStr + ". Must be a positive integer")
		}
	}

	var MqttMessageCleaningTimeoutStr string
	MqttMessageCleaningTimeoutStr, exists = os.LookupEnv("MQTT_MESSAGE_CLEANING_TIMEOUT")
	if exists {
		var err error
		MqttMessageCleaningTimeout, err = strconv.Atoi(MqttMessageCleaningTimeoutStr)
		if err != nil || MqttMessageCleaningTimeout <= 0 {
			return errors.New("invalid value for MQTT_MESSAGE_CLEANING_TIMEOUT: " + MqttMessageCleaningTimeoutStr + ". Must be a positive integer")
		}
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

	ProximityRealtimeDataTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_REALTIME_DATA_TOPIC")
	if !exists {
		ProximityRealtimeDataTopic = kafka.PROXIMITY_FOG_HUB_REALTIME_DATA_TOPIC
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

	KafkaPublishTimeoutStr, exists := os.LookupEnv("KAFKA_PUBLISH_TIMEOUT")
	if exists {
		var err error
		KafkaPublishTimeout, err = strconv.Atoi(KafkaPublishTimeoutStr)
		if err != nil || KafkaPublishTimeout <= 0 {
			return errors.New("invalid value for KAFKA_PUBLISH_TIMEOUT: " + KafkaPublishTimeoutStr + ". Must be a positive integer")
		}
	}

	/* ----- POSTGRESQL DATABASE SETTINGS ----- */

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
		PostgresHost = "localhost"
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
