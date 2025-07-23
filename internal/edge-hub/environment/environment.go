package environment

import (
	"SensorContinuum/configs/kafka"
	"SensorContinuum/configs/mosquitto"
	"errors"
	"github.com/google/uuid"
	"os"
	"strconv"
	"time"
)

// I valori validi per OperationMode sono:
// - "loop": per eseguire in un ciclo continuo di aggregazione|pulizia dei dati. (default)
// - "once": per eseguire una singola iterazione di aggregazione|pulizia dei dati.

// OperationMode specifica la modalità di funzionamento del servizio.
var OperationMode string

var BuildingID string
var FloorID string
var HubID string

var MosquittoProtocol string
var MosquittoBroker string
var MosquittoPort string
var SensorDataTopic string

var KafkaBroker string
var KafkaPort string
var EdgeHubTopic string
var EdgeHubTopicPartition string

var RedisAddress string
var RedisPort string

var HistoryWindowSize int = 25
var FilteringMinSamples int = 5
var FilteringStdDevFactor float64 = 5
var FilteringMinTreshold float64 = 0.0
var FilteringMaxTreshold float64 = 60.0

var UnhealthySensorTimeout time.Duration = 5 * time.Minute

func SetupEnvironment() error {

	var exists bool

	OperationMode, exists = os.LookupEnv("OPERATION_MODE")
	if !exists {
		OperationMode = "loop"
	}

	BuildingID, exists = os.LookupEnv("BUILDING_ID")
	if !exists {
		return errors.New("environment variable BUILDING_ID not set")
	}

	FloorID, exists = os.LookupEnv("FLOOR_ID")
	if !exists {
		return errors.New("environment variable FLOOR_ID not set")
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

	SensorDataTopic = "$share/edge-hub-filtering/" + BuildingID + "/" + FloorID + "/"

	KafkaBroker, exists = os.LookupEnv("KAFKA_BROKER_ADDRESS")
	if !exists {
		KafkaBroker = kafka.BROKER
	}

	KafkaPort, exists = os.LookupEnv("KAFKA_BROKER_PORT")
	if !exists {
		KafkaPort = kafka.PORT
	}

	EdgeHubTopic, exists = os.LookupEnv("KAFKA_EDGE_HUB_TOPIC")
	if !exists {
		EdgeHubTopic = kafka.EDGE_HUB_TOPIC + "_" + BuildingID
	}

	EdgeHubTopicPartition, exists = os.LookupEnv("KAFKA_EDGE_HUB_TOPIC_PARTITION")
	if !exists {
		EdgeHubTopicPartition = FloorID
	}

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

	FilteringMinTresholdStr, exists := os.LookupEnv("FILTERING_MIN_TRESHOLD")
	if exists {
		var err error
		FilteringMinTreshold, err = strconv.ParseFloat(FilteringMinTresholdStr, 64)
		if err != nil {
			return errors.New("invalid value for FILTERING_MIN_TRESHOLD: " + FilteringMinTresholdStr)
		}
	}

	FilteringMaxTresholdStr, exists := os.LookupEnv("FILTERING_MAX_TRESHOLD")
	if exists {
		var err error
		FilteringMaxTreshold, err = strconv.ParseFloat(FilteringMaxTresholdStr, 64)
		if err != nil {
			return errors.New("invalid value for FILTERING_MAX_TRESHOLD: " + FilteringMaxTresholdStr)
		}
	}

	return nil

}
