package environment

import (
	"SensorContinuum/configs/mosquitto"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// OperationMode: I valori validi sono:
// - "loop": per eseguire in un ciclo continuo di aggregazione|pulizia dei dati. (default)
// - "once": per eseguire una singola iterazione di aggregazione|pulizia dei dati.
type OperationModeType string

const (
	OperationModeLoop OperationModeType = "loop"
	OperationModeOnce OperationModeType = "once"
)

// OperationMode specifica la modalità di funzionamento del servizio.
var OperationMode OperationModeType

var ServiceMode types.Service

var EdgeMacrozone string
var EdgeZone string
var HubID string

// Queste impostazioni sono utilizzate per la connessione tra SensorAgent e EdgeHub.

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

var RedisAddress string
var RedisPort string

var HistoryWindowSize int = 100
var FilteringMinSamples int = 5
var FilteringStdDevFactor float64 = 5

var UnhealthySensorTimeout time.Duration = 5 * time.Minute
var RegistrationSensorTimeout time.Duration = 6 * time.Hour

var HealthzServer bool = false
var HealthzServerPort string = ":"

func SetupEnvironment() error {

	var exists bool

	/* ----- OPERATION MODE ----- */

	var OperationModeStr string
	OperationModeStr, exists = os.LookupEnv("OPERATION_MODE")
	if !exists {
		OperationMode = OperationModeLoop
	} else {
		switch OperationModeStr {
		case string(OperationModeLoop):
			OperationMode = OperationModeLoop
		case string(OperationModeOnce):
			OperationMode = OperationModeOnce
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

	SensorDataTopic = "$share/edge-hub-data/sensor-data/" + EdgeMacrozone + "/" + EdgeZone
	FilteredDataTopic = "filtered-data/" + EdgeMacrozone + "/" + EdgeZone
	HubConfigurationTopic = "configuration/hub/" + EdgeMacrozone + "/" + EdgeZone
	SensorConfigurationTopic = "configuration/sensor/" + EdgeMacrozone + "/" + EdgeZone
	HeartbeatTopic = "heartbeat/" + EdgeMacrozone + "/" + EdgeZone

	/* ----- REDIS CACHE SETTINGS ----- */

	RedisAddress, exists = os.LookupEnv("REDIS_ADDRESS")
	if !exists {
		RedisAddress = "localhost"
	}

	RedisPort, exists = os.LookupEnv("REDIS_PORT")
	if !exists {
		RedisPort = "6379"
	}

	HistoryWindowSizeStr, exists := os.LookupEnv("HISTORY_WINDOW_SIZE")
	if exists {
		var err error
		HistoryWindowSize, err = strconv.Atoi(HistoryWindowSizeStr)
		if err != nil {
			return errors.New("invalid value for HISTORY_WINDOW_SIZE: " + HistoryWindowSizeStr)
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

	/* ----- LOGGER SETTINGS ----- */

	if err := logger.LoadLoggerFromEnv(); err != nil {
		return err
	}

	return nil

}
