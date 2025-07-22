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
var IntermediateDataTopic string

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

	ProximityDataTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_TOPIC")
	if !exists {
		ProximityDataTopic = kafka.PROXIMITY_FOG_HUB_TOPIC + "_" + BuildingID
	}

	IntermediateDataTopic, exists = os.LookupEnv("KAFKA_INTERMEDIATE_FOG_HUB_TOPIC")
	if !exists {
		IntermediateDataTopic = kafka.INTERMEDIATE_FOG_HUB_TOPIC
	}

	return nil

}
