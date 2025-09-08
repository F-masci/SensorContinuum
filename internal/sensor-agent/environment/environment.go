package environment

import (
	"SensorContinuum/configs/mosquitto"
	"SensorContinuum/pkg/logger"
	"errors"
	"os"
	"strconv"

	"github.com/google/uuid"
)

type Location string

const (
	Indoor  Location = "indoor"
	Outdoor Location = "outdoor"
)

// SensorLocation identifica se il sensore è installato in un edificio o all'esterno.
// I valori possibili sono "indoor" o "outdoor".
var SensorLocation Location

type Type string

const (
	Temperature Type = "temperature"
	Humidity    Type = "humidity"
	Pressure    Type = "pressure"
)

// SensorType identifica il tipo di sensore, ad esempio "temperature", "humidity", "pressure", ecc.
var SensorType Type

// SimulationValueColumn è il nome della colonna utilizzata per la simulazione del sensore.
var SimulationValueColumn string

// SimulationTimestampColumn è il nome della colonna utilizzata per il timestamp del sensore.
var SimulationTimestampColumn string

// SimulationSeparator è il separatore utilizzato per i valori simulati del sensore.
var SimulationSeparator rune

type Reference string

// I possibili sensori di riferimento per la simulazione sono:
const (
	BMP280          Reference = "bmp280"
	DHT22           Reference = "dht22"
	DS18B20         Reference = "ds18b20"
	HPM_SENSOR      Reference = "hpm_sensor"
	HTU21D          Reference = "htu21d"
	LAERM_SENSOR    Reference = "laerm_sensor"
	PMS5003         Reference = "pms5003"
	PMS7003         Reference = "pms7003"
	PPD42NS         Reference = "ppd42ns"
	RADIATION_SBM19 Reference = "radiation_sbm-19"
	RADIATION_SBM20 Reference = "radiation_sbm-20"
	RADIATION_SI22G Reference = "radiation_si22g"
	SCD30           Reference = "scd30"
	SDS011          Reference = "sds011"
	SHT30           Reference = "sht30"
	SHT31           Reference = "sht31"
	SPS30           Reference = "sps30"
)

// SimulationSensorReference è il riferimento al sensore utilizzato (nome) per la simulazione.
var SimulationSensorReference Reference

// SimulationTimestampFormat è il formato del timestamp utilizzato per la simulazione.
var SimulationTimestampFormat string

// SimulationOffsetDay è il numero di giorni da sottrarre alla data corrente per ottenere la data di simulazione.
// Ad esempio, se oggi è 2024-06-15 e SimulationOffsetDay è 2, la data di simulazione sarà 2024-06-13.
var SimulationOffsetDay int = 2

type IdGenerator string

const (
	UUID     IdGenerator = "uuid"
	Hostname IdGenerator = "hostname"
)

var EdgeMacrozone string
var EdgeZone string
var SensorIdGenerator IdGenerator
var SensorId string

var MqttBrokerProtocol string
var MqttBrokerAddress string
var MqttBrokerPort string
var DataTopic string
var ConfigurationTopic string

var MaxReconnectionInterval int = 10 // in seconds
var MaxReconnectionTimeout int = 10  // in seconds
var MaxReconnectionAttempts int = 10
var MessagePublishTimeout int = 5 // in seconds

var HealthzServer bool = false
var HealthzServerPort string = ":"

func SetupEnvironment() error {

	/* ----- SIMULATION SETTINGS ----- */

	SensorLocationStr, exists := os.LookupEnv("SENSOR_LOCATION")
	if !exists {
		SensorLocation = Indoor
	} else {
		switch SensorLocationStr {
		case string(Indoor):
			SensorLocation = Indoor
		case string(Outdoor):
			SensorLocation = Outdoor
		default:
			return errors.New("invalid SENSOR_LOCATION value, must be 'indoor' or 'outdoor'")
		}
	}

	var SensorTypeStr string
	SensorTypeStr, exists = os.LookupEnv("SENSOR_TYPE")
	if !exists {
		SensorType = Temperature
	} else {
		switch SensorTypeStr {
		case string(Temperature):
			SensorType = Temperature
		case string(Humidity):
			SensorType = Humidity
		case string(Pressure):
			SensorType = Pressure
		default:
			return errors.New("invalid SENSOR_TYPE value, must be 'temperature', 'humidity', or 'pressure'")
		}
	}

	SimulationValueColumn, exists = os.LookupEnv("SIMULATION_VALUE_COLUMN")
	if !exists {
		SimulationValueColumn = string(SensorType)
	}

	SimulationTimestampColumn, exists = os.LookupEnv("SIMULTION_TIMESTAMP_COLUMN")
	if !exists {
		SimulationTimestampColumn = "timestamp"
	}

	var SimulationSeparatorStr string
	SimulationSeparatorStr, exists = os.LookupEnv("SIMULATION_SEPARATOR")
	if !exists {
		SimulationSeparator = ';'
	} else {
		SimulationSeparator = []rune(SimulationSeparatorStr)[0]
	}

	var SimulationSensorReferenceStr string
	SimulationSensorReferenceStr, exists = os.LookupEnv("SIMULATION_SENSOR_REFERENCE")
	if !exists {
		SimulationSensorReference = BMP280
	} else {
		switch SimulationSensorReferenceStr {
		case string(BMP280):
			SimulationSensorReference = BMP280
		case string(DHT22):
			SimulationSensorReference = DHT22
		case string(DS18B20):
			SimulationSensorReference = DS18B20
		case string(HPM_SENSOR):
			SimulationSensorReference = HPM_SENSOR
		case string(HTU21D):
			SimulationSensorReference = HTU21D
		case string(LAERM_SENSOR):
			SimulationSensorReference = LAERM_SENSOR
		case string(PMS5003):
			SimulationSensorReference = PMS5003
		case string(PMS7003):
			SimulationSensorReference = PMS7003
		case string(PPD42NS):
			SimulationSensorReference = PPD42NS
		case string(RADIATION_SBM19):
			SimulationSensorReference = RADIATION_SBM19
		case string(RADIATION_SBM20):
			SimulationSensorReference = RADIATION_SBM20
		case string(RADIATION_SI22G):
			SimulationSensorReference = RADIATION_SI22G
		case string(SCD30):
			SimulationSensorReference = SCD30
		case string(SDS011):
			SimulationSensorReference = SDS011
		case string(SHT30):
			SimulationSensorReference = SHT30
		case string(SHT31):
			SimulationSensorReference = SHT31
		case string(SPS30):
			SimulationSensorReference = SPS30
		default:
			return errors.New("invalid value for SIMULATION_SENSOR_REFERENCE: " + SimulationSensorReferenceStr + ". Must be one of the predefined sensor references")
		}
	}

	SimulationTimestampFormat, exists = os.LookupEnv("SIMULATION_TIMESTAMP_FORMAT")
	if !exists {
		SimulationTimestampFormat = "2006-01-02T15:04:05"
	}

	var SimulationOffsetDayStr string
	SimulationOffsetDayStr, exists = os.LookupEnv("SIMULATION_OFFSET_DAY")
	if exists {
		var err error
		SimulationOffsetDay, err = strconv.Atoi(SimulationOffsetDayStr)
		if err != nil || SimulationOffsetDay < 0 {
			return errors.New("invalid value for SIMULATION_OFFSET_DAY: " + SimulationOffsetDayStr + ". Must be a non-negative integer")
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

	var SensorIdGeneratorStr string
	SensorIdGeneratorStr, exists = os.LookupEnv("SENSOR_ID_GENERATOR")
	if !exists {
		SensorIdGenerator = UUID
	} else {
		switch SensorIdGeneratorStr {
		case string(UUID):
			SensorIdGenerator = UUID
		case string(Hostname):
			SensorIdGenerator = Hostname
		default:
			return errors.New("invalid value for SENSOR_ID_GENERATOR: " + SensorIdGeneratorStr + ". Must be 'uuid' or 'hostname'")
		}
	}

	SensorId, exists = os.LookupEnv("SENSOR_ID")
	if !exists {
		switch SensorIdGenerator {
		case UUID:
			SensorId = uuid.New().String()
		case Hostname:
			var hostname string
			var err error

			// Prova a ottenere il nome host dall'ambiente
			hostname, exists = os.LookupEnv("HOSTNAME")
			if !exists {
				// Se non è impostato, ottieni il nome host del sistema
				hostname, err = os.Hostname()
				if err != nil {
					return errors.New("failed to get hostname: " + err.Error())
				}
			}
			SensorId = hostname + "-" + EdgeMacrozone + "-" + EdgeZone
		default:
			return errors.New("invalid SENSOR_ID_GENERATOR: " + SensorIdGeneratorStr + ". Must be 'uuid' or 'hostname'")
		}
	}

	/* ----- MQTT BROKER SETTINGS ----- */

	MqttBrokerProtocol, exists = os.LookupEnv("MQTT_BROKER_PROTOCOL")
	if !exists {
		MqttBrokerProtocol = mosquitto.PROTOCOL
	}

	MqttBrokerAddress, exists = os.LookupEnv("MQTT_BROKER_ADDRESS")
	if !exists {
		MqttBrokerAddress = mosquitto.BROKER
	}

	MqttBrokerPort, exists = os.LookupEnv("MQTT_BROKER_PORT")
	if !exists {
		MqttBrokerPort = mosquitto.PORT
	}

	DataTopic = "sensor-data/" + EdgeMacrozone + "/" + EdgeZone
	ConfigurationTopic = "configuration/sensor/" + EdgeMacrozone + "/" + EdgeZone

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

	var MessagePublishTimeoutStr string
	MessagePublishTimeoutStr, exists = os.LookupEnv("MESSAGE_PUBLISH_TIMEOUT")
	if exists {
		var err error
		MessagePublishTimeout, err = strconv.Atoi(MessagePublishTimeoutStr)
		if err != nil || MessagePublishTimeout <= 0 {
			return errors.New("invalid value for MESSAGE_PUBLISH_TIMEOUT: " + MessagePublishTimeoutStr + ". Must be a positive integer")
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

	if err := logger.LoadLoggerFromEnv(); err != nil {
		return err
	}

	return nil

}
