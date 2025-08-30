package environment

import (
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

// ServiceMode specifica il tipo di servizio Edge Hub in esecuzione.
var ServiceMode types.Service

var EdgeMacrozone string
var EdgeZone string
var HubID string

// Queste impostazioni sono utilizzate per la connessione tra:
// - SensorAgent -> EdgeHub
// - EdgeHub -> ProximityFogHub

// MqttSensorBrokerProtocol specifica il protocollo (es. "tcp", "ws") tra SensorAgent ed EdgeHub.
var MqttSensorBrokerProtocol string

// MqttSensorBrokerAddress specifica l'indirizzo del broker MQTT per la comunicazione tra SensorAgent ed EdgeHub.
var MqttSensorBrokerAddress string

// MqttSensorBrokerPort specifica la porta del broker MQTT per la comunicazione tra SensorAgent ed EdgeHub.
var MqttSensorBrokerPort string

// Queste impostazioni sono utilizzate per la connessione tra EdgeHub e Proximity Hub.

// MqttHubBrokerProtocol specifica il protocollo (es. "tcp", "ws") tra EdgeHub e Proximity Hub.
var MqttHubBrokerProtocol string

// MqttHubBrokerAddress specifica l'indirizzo del broker MQTT per la comunicazione tra EdgeHub e Proximity Hub.
var MqttHubBrokerAddress string

// MqttHubBrokerPort specifica la porta del broker MQTT per la comunicazione tra EdgeHub e Proximity Hub.
var MqttHubBrokerPort string

// SensorDataTopic specifica il topic MQTT per i dati dei sensori.
var SensorDataTopic string

// FilteredDataTopic specifica il topic MQTT per i dati filtrati.
var FilteredDataTopic string

// HubConfigurationTopic specifica il topic MQTT per i messaggi di configurazione del hub.
var HubConfigurationTopic string

// SensorConfigurationTopic specifica il topic MQTT per i messaggi di configurazione dei sensori.
var SensorConfigurationTopic string

// HeartbeatTopic specifica il topic MQTT per i messaggi di heartbeat del hub.
var HeartbeatTopic string

// Queste impostazioni controllano il comportamento della riconnessione al broker MQTT.

// MaxReconnectionInterval specifica l'intervallo massimo tra i tentativi di riconnessione in secondi.
var MaxReconnectionInterval int = 10

// MaxReconnectionTimeout specifica il timeout massimo per ogni tentativo di riconnessione in secondi.
var MaxReconnectionTimeout int = 10

// MaxReconnectionAttempts specifica il numero massimo di tentativi di riconnessione.
var MaxReconnectionAttempts int = 10

// MaxSubscriptionTimeout specifica il timeout per la sottoscrizione ai topic in secondi.
var MaxSubscriptionTimeout int = 5

// MessagePublishTimeout specifica il timeout per la pubblicazione dei messaggi in secondi.
var MessagePublishTimeout int = 5

// MessagePublishAttempts specifica il numero di tentativi di pubblicazione dei messaggi.
var MessagePublishAttempts int = 3

// MessageCleaningTimeout specifica il timeout per la pulizia dei messaggi in secondi.
var MessageCleaningTimeout int = MessagePublishTimeout

var RedisAddress string
var RedisPort string

// AggregationInterval specifica l'intervallo di tempo per l'aggregazione dei dati.
const AggregationInterval = time.Minute

// AggregationFetchOffset è il tempo extra per recuperare i dati dai sensori,
// in modo da includere eventuali ritardi nella ricezione dei messaggi.
// Specifica l'offest negativo di tempo rispetto all'istante corrente
// per recuperare i dati aggregati.
const AggregationFetchOffset = -2 * time.Minute

const LeaderKey = "edge-hub-leader"
const LeaderTTL = 70 * time.Second

const HistoryWindowSize int = 100
const FilteringMinSamples int = 5
const FilteringStdDevFactor float64 = 3

const UnhealthySensorTimeout = timeouts.IsAliveSensorTimeout
const RegistrationSensorTimeout = 6 * time.Hour

const HeartbeatInterval = timeouts.HeartbeatInterval

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
		ServiceMode = types.EdgeHubService
	} else {
		switch ServiceModeStr {
		case string(types.EdgeHubService):
			ServiceMode = types.EdgeHubService
		case string(types.EdgeHubFilterService):
			ServiceMode = types.EdgeHubFilterService
		case string(types.EdgeHubAggregatorService):
			ServiceMode = types.EdgeHubAggregatorService
		case string(types.EdgeHubCleanerService):
			ServiceMode = types.EdgeHubCleanerService
		case string(types.EdgeHubConfigurationService):
			ServiceMode = types.EdgeHubConfigurationService
		default:
			return errors.New("invalid value for SERVICE_MODE: " + ServiceModeStr + ". Valid values are 'edge-hub', 'edge-hub-filter', 'edge-hub-aggregator', 'edge-hub-cleaner' or 'edge-hub-configuration'.")
		}
	}

	/* ----- ENVIRONMENT SETTINGS ----- */

	EdgeMacrozone, exists = os.LookupEnv("EDGE_MACROZONE")
	if !exists {
		return errors.New("environment variable EDGE_MACROZONE not set")
	}

	EdgeZone, exists = os.LookupEnv("EDGE_ZONE")
	if !exists {
		return errors.New("environment variable EDGE_ZONE not set")
	}

	HubID, exists = os.LookupEnv("HUB_ID")
	if !exists {
		HubID = uuid.New().String()
	}

	/* ----- MQTT BROKER SETTINGS ----- */

	// La configurazione del broker MQTT è divisa in due parti:
	// 1. MqttSensorBroker: per la comunicazione tra SensorAgent e EdgeHub.
	// SensorAgent ---> MqttSensorBroker ---> EdgeHub
	// 2. MqttHubBroker: per la comunicazione tra EdgeHub e Proximity Hub.
	// EdgeHub ---> MqttHubBroker ---> Proximity Hub

	// Recupera le configurazioni comuni una sola volta
	commonProtocol, exists := os.LookupEnv("MQTT_BROKER_PROTOCOL")
	if !exists {
		commonProtocol = mosquitto.PROTOCOL
	}
	commonBroker, exists := os.LookupEnv("MQTT_BROKER_ADDRESS")
	if !exists {
		commonBroker = mosquitto.BROKER
	}
	commonPort, exists := os.LookupEnv("MQTT_BROKER_PORT")
	if !exists {
		commonPort = mosquitto.PORT
	}

	// Sensor Broker
	MqttSensorBrokerProtocol, exists = os.LookupEnv("MQTT_SENSOR_BROKER_PROTOCOL")
	if !exists {
		MqttSensorBrokerProtocol = commonProtocol
	}
	MqttSensorBrokerAddress, exists = os.LookupEnv("MQTT_SENSOR_BROKER_ADDRESS")
	if !exists {
		MqttSensorBrokerAddress = commonBroker
	}
	MqttSensorBrokerPort, exists = os.LookupEnv("MQTT_SENSOR_BROKER_PORT")
	if !exists {
		MqttSensorBrokerPort = commonPort
	}

	// Hub Broker
	MqttHubBrokerProtocol, exists = os.LookupEnv("MQTT_HUB_BROKER_PROTOCOL")
	if !exists {
		MqttHubBrokerProtocol = commonProtocol
	}
	MqttHubBrokerAddress, exists = os.LookupEnv("MQTT_HUB_BROKER_ADDRESS")
	if !exists {
		MqttHubBrokerAddress = commonBroker
	}
	MqttHubBrokerPort, exists = os.LookupEnv("MQTT_HUB_BROKER_PORT")
	if !exists {
		MqttHubBrokerPort = commonPort
	}

	SensorDataTopic = "$share/edge-hub_" + EdgeMacrozone + "_" + EdgeZone + "/sensor-data/" + EdgeMacrozone + "/" + EdgeZone
	FilteredDataTopic = "filtered-data/" + EdgeMacrozone + "/" + EdgeZone
	HubConfigurationTopic = "configuration/hub/" + EdgeMacrozone + "/" + EdgeZone
	SensorConfigurationTopic = "configuration/sensor/" + EdgeMacrozone + "/" + EdgeZone
	HeartbeatTopic = "heartbeat/" + EdgeMacrozone + "/" + EdgeZone

	var MaxReconnectionIntervalStr string
	MaxReconnectionIntervalStr, exists = os.LookupEnv("MAX_RECONNECTION_INTERVAL")
	if exists {
		var err error
		MaxReconnectionInterval, err = strconv.Atoi(MaxReconnectionIntervalStr)
		if err != nil || MaxReconnectionInterval <= 0 {
			return errors.New("invalid value for MAX_RECONNECTION_INTERVAL: " + MaxReconnectionIntervalStr + ". Must be a positive integer")
		}
	}

	var MaxReconnectionTimeoutStr string
	MaxReconnectionTimeoutStr, exists = os.LookupEnv("MAX_RECONNECTION_TIMEOUT")
	if exists {
		var err error
		MaxReconnectionTimeout, err = strconv.Atoi(MaxReconnectionTimeoutStr)
		if err != nil || MaxReconnectionTimeout <= 0 {
			return errors.New("invalid value for MAX_RECONNECTION_TIMEOUT: " + MaxReconnectionTimeoutStr + ". Must be a positive integer")
		}
	}

	var MaxReconnectionAttemptsStr string
	MaxReconnectionAttemptsStr, exists = os.LookupEnv("MAX_RECONNECTION_ATTEMPTS")
	if exists {
		var err error
		MaxReconnectionAttempts, err = strconv.Atoi(MaxReconnectionAttemptsStr)
		if err != nil || MaxReconnectionAttempts <= 0 {
			return errors.New("invalid value for MAX_RECONNECTION_ATTEMPTS: " + MaxReconnectionAttemptsStr + ". Must be a positive integer")
		}
	}

	var MaxSubscriptionTimeoutStr string
	MaxSubscriptionTimeoutStr, exists = os.LookupEnv("MAX_SUBSCRIPTION_TIMEOUT")
	if exists {
		var err error
		MaxSubscriptionTimeout, err = strconv.Atoi(MaxSubscriptionTimeoutStr)
		if err != nil || MaxSubscriptionTimeout <= 0 {
			return errors.New("invalid value for MAX_SUBSCRIPTION_TIMEOUT: " + MaxSubscriptionTimeoutStr + ". Must be a positive integer")
		}
	}

	var MessagePublishTimeoutStr string
	MessagePublishTimeoutStr, exists = os.LookupEnv("MESSAGE_PUBLISH_TIMEOUT")
	if exists {
		var err error
		MessagePublishTimeout, err = strconv.Atoi(MessagePublishTimeoutStr)
		if err != nil || MessagePublishTimeout <= 0 {
			return errors.New("invalid value for MESSAGE_PUBLISH_TIMEOUT: " + MessagePublishTimeoutStr + ". Must be a positive integer")
		}
	}

	var MessagePublishAttemptsStr string
	MessagePublishAttemptsStr, exists = os.LookupEnv("MESSAGE_PUBLISH_ATTEMPTS")
	if exists {
		var err error
		MessagePublishAttempts, err = strconv.Atoi(MessagePublishAttemptsStr)
		if err != nil || MessagePublishAttempts <= 0 {
			return errors.New("invalid value for MESSAGE_PUBLISH_ATTEMPTS: " + MessagePublishAttemptsStr + ". Must be a positive integer")
		}
	}

	var MessageCleaningTimeoutStr string
	MessageCleaningTimeoutStr, exists = os.LookupEnv("MESSAGE_CLEANING_TIMEOUT")
	if exists {
		var err error
		MessageCleaningTimeout, err = strconv.Atoi(MessageCleaningTimeoutStr)
		if err != nil || MessageCleaningTimeout <= 0 {
			return errors.New("invalid value for MESSAGE_CLEANING_TIMEOUT: " + MessageCleaningTimeoutStr + ". Must be a positive integer")
		}
	}

	/* ----- REDIS CACHE SETTINGS ----- */

	RedisAddress, exists = os.LookupEnv("REDIS_ADDRESS")
	if !exists {
		RedisAddress = "localhost"
	}

	RedisPort, exists = os.LookupEnv("REDIS_PORT")
	if !exists {
		RedisPort = "6379"
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
