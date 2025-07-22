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
var EdgeHubTopic string
var ProximityDataTopic string

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

	ProximityDataTopic, exists = os.LookupEnv("KAFKA_PROXIMITY_FOG_HUB_TOPIC")
	if !exists {
		ProximityDataTopic = kafka.PROXIMITY_FOG_HUB_TOPIC + "_" + BuildingID
	}

	return nil

}
