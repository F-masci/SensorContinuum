package environment

import (
	"SensorContinuum/configs/mosquitto"
	"errors"
	"github.com/google/uuid"
	"os"
)

// SensorLocation identifica se il sensore è installato in un edificio o all'esterno.
// I valori possibili sono "building" o "outdoor".
var SensorLocation string

// SensorType identifica il tipo di sensore, ad esempio "temperature", "humidity", "humidity", ecc.
var SensorType string

// SimulationValueColumn è il nome della colonna utilizzata per la simulazione del sensore.
var SimulationValueColumn string

// SimulationTimestampColumn è il nome della colonna utilizzata per il timestamp del sensore.
var SimulationTimestampColumn string

// SimulationSeparator è il separatore utilizzato per i valori simulati del sensore.
var SimulationSeparator rune

// I possibili sensori di riferimento per la simulazione sono:
// - "bmp280"
// - "dht22"
// - "ds18b20"
// - "hpm_sensor"
// - "htu21d"
// - "laerm_sensor"
// - "pms5003"
// - "pms7003"
// - "ppd42ns"
// - "radiation_sbm-19"
// - "radiation_sbm-20"
// - "radiation_si22g"
// - "scd30"
// - "sds011"
// - "sht30"
// - "sht31"
// - "sps30"

// SimulationSensorReference è il riferimento al sensore utilizzato (nome) per la simulazione.
var SimulationSensorReference string

// SimulationTimestampFormat è il formato del timestamp utilizzato per la simulazione.
var SimulationTimestampFormat string

var BuildingID string
var FloorID string
var SensorID string

var MosquittoProtocol string
var MosquittoBroker string
var MosquittoPort string
var BaseTopic string

func SetupEnvironment() error {

	var exists bool

	SensorLocation, exists = os.LookupEnv("SENSOR_LOCATION")
	if !exists {
		SensorLocation = "building"
	}

	SensorType, exists = os.LookupEnv("SENSOR_TYPE")
	if !exists {
		SensorType = "temperature"
	}

	SimulationValueColumn, exists = os.LookupEnv("SIMULATION_VALUE_COLUMN")
	if !exists {
		SimulationValueColumn = SensorType
	}

	SimulationTimestampColumn, exists = os.LookupEnv("SIMULTION_TIMESTAMP_COLUMN")
	if !exists {
		SimulationTimestampColumn = "timestamp"
	}

	SimulationSeparatorStr, exists := os.LookupEnv("SIMULATION_SEPARATOR")
	if !exists {
		SimulationSeparator = ';'
	} else {
		SimulationSeparator = []rune(SimulationSeparatorStr)[0]
	}

	SimulationSensorReference, exists = os.LookupEnv("SIMULATION_SENSOR_REFERENCE")
	if !exists {
		SimulationSensorReference = "bmp280"
	}

	SimulationTimestampFormat, exists = os.LookupEnv("SIMULATION_TIMESTAMP_FORMAT")
	if !exists {
		SimulationTimestampFormat = "2006-01-02T15:04:05"
	}

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

	BaseTopic = "sensor-data/" + BuildingID + "/" + FloorID

	return nil

}
