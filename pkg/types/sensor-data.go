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
	Timestamp     int64   `json:"timestamp"`
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

type SensorDataBatch struct {
	SensorData []SensorData `json:"sensor_data"`
	counter    int
}

// NewSensorDataBatch crea un nuovo batch di dati sensori
func NewSensorDataBatch() SensorDataBatch {
	return SensorDataBatch{
		SensorData: make([]SensorData, 0),
		counter:    0,
	}
}

// AddSensorData aggiunge un nuovo dato sensore al batch e incrementa il contatore
func (sdb *SensorDataBatch) AddSensorData(data SensorData) {
	sdb.SensorData = append(sdb.SensorData, data)
	sdb.counter++
}

// Count restituisce il numero di dati sensori nel batch
func (sdb SensorDataBatch) Count() int {
	return sdb.counter
}

// Clear resetta il batch di dati sensori
func (sdb *SensorDataBatch) Clear() {
	sdb.SensorData = make([]SensorData, 0)
	sdb.counter = 0
}

// AggregatedStats contiene i dati statistici calcolati ogni tot minuti dal Proximity-Fog-Hub
// e inviati tramite kafka all' intermediate-fog-hub
type AggregatedStats struct {
	Timestamp int64   `json:"timestamp"`
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
