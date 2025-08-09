package types

import (
	"encoding/json"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/segmentio/kafka-go"
)

type SensorData struct {
	EdgeMacrozone string  `json:"macrozone"`
	EdgeZone      string  `json:"zone"`
	SensorID      string  `json:"sensor_id"`
	Timestamp     string  `json:"timestamp"`
	Type          string  `json:"type"`
	Data          float64 `json:"data"`
}

func CreateSensorDataFromMQTT(msg MQTT.Message) (SensorData, error) {
	var data SensorData
	err := json.Unmarshal(msg.Payload(), &data)
	return data, err
}

func CreateSensorDataFromKafka(msg kafka.Message) (SensorData, error) {
	var data SensorData
	err := json.Unmarshal(msg.Value, &data)
	return data, err
}
