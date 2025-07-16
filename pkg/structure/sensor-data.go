package structure

import (
	"encoding/json"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type SensorData struct {
	BuildingID string  `json:"building_id"`
	FloorID    string  `json:"floor_id"`
	SensorID   string  `json:"sensor_id"`
	Timestamp  string  `json:"timestamp"`
	Data       float64 `json:"data"`
}

func CreateSensorDataFromMQTT(msg MQTT.Message) (SensorData, error) {
	var data SensorData
	err := json.Unmarshal(msg.Payload(), &data)
	return data, err
}
