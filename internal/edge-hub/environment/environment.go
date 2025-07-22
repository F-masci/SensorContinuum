package environment

import (
	"SensorContinuum/configs/kafka"
	"SensorContinuum/configs/mosquitto"
	"SensorContinuum/pkg/logger"
	"errors"
	"github.com/google/uuid"
	"os"
	"regexp"
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

	SensorDataTopic = BuildingID + "/" + FloorID + "/"

	KafkaBroker, exists = os.LookupEnv("KAFKA_ADDRESS")
	if !exists {
		KafkaBroker = kafka.BROKER
	}

	KafkaPort, exists = os.LookupEnv("KAFKA_PORT")
	if !exists {
		KafkaPort = kafka.PORT
	}

	EdgeHubTopic, exists = os.LookupEnv("KAFKA_EDGE_HUB_TOPIC")
	if !exists {
		EdgeHubTopic = kafka.EDGE_HUB_TOPIC + "_" + BuildingID
	}

	EdgeHubTopicPartition, exists := os.LookupEnv("KAFKA_EDGE_HUB_TOPIC_PARTITION")
	if !exists {
		EdgeHubTopicPartition = FloorID
	}

	return nil

}
