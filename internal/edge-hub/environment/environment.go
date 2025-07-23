package environment

import (
	"SensorContinuum/configs/kafka"
	"SensorContinuum/configs/mosquitto"
	"errors"
	"github.com/google/uuid"
	"os"
	"strconv"
)

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

var HistoryWindowSize int = 10
var FilteringMinSamples int = 5
var FilteringStdDevFactor float64 = 5

func SetupEnvironment() error {

	var exists bool

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

	HistoryWindowSizeStr, exists := os.LookupEnv("HISTORY_WINDOW_SIZE")
	if !exists {
		var err error
		HistoryWindowSize, err = strconv.Atoi(HistoryWindowSizeStr)
		if err != nil {
			HistoryWindowSize = 10
		}
	}

	return nil

}
