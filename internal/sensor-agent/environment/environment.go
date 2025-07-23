package environment

import (
	"errors"
	"github.com/google/uuid"
	"os"

	"SensorContinuum/configs/mosquitto"
)

var BuildingID string
var FloorID string
var SensorID string

var MosquittoProtocol string
var MosquittoBroker string
var MosquittoPort string
var BaseTopic string

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

	SensorID, exists = os.LookupEnv("SENSOR_ID")
	if !exists {
		SensorID = uuid.New().String()
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

	BaseTopic = BuildingID + "/" + FloorID + "/"

	return nil

}
