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

// AggregatedStats contiene i dati statistici calcolati ogni tot minuti dal Proximity-Fog-Hub
// e inviati tramite kafka all' intermediate-fog-hub
type AggregatedStats struct {
	Timestamp string  `json:"timestamp"`
	Macrozone string  `json:"macrozone"`
	Type      string  `json:"type"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Avg       float64 `json:"avg"`
}

// CreateAggregatedStatsFromKafka deserializza un messaggio Kafka in AggregatedStats
func CreateAggregatedStatsFromKafka(msg kafka.Message) (AggregatedStats, error) {
	var stats AggregatedStats
	err := json.Unmarshal(msg.Value, &stats)
	return stats, err
}
