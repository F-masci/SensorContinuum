package types

import (
	"encoding/json"
	"time"

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

	KafkaMsg kafka.Message `json:"-"`
	MQTTMsg  MQTT.Message  `json:"-"`
}

func CreateSensorDataFromMQTT(msg MQTT.Message) (SensorData, error) {
	var data SensorData
	err := json.Unmarshal(msg.Payload(), &data)
	data.MQTTMsg = msg
	return data, err
}

func CreateSensorDataFromKafka(msg kafka.Message) (SensorData, error) {
	var data SensorData
	err := json.Unmarshal(msg.Value, &data)
	data.KafkaMsg = msg
	return data, err
}

type SensorDataBatch struct {
	engine *BatchEngine[SensorData]
}

func NewSensorDataBatch(maxCount int, timeout time.Duration, save func(*SensorDataBatch) error) (*SensorDataBatch, error) {
	sdb := &SensorDataBatch{}
	var err error
	sdb.engine, err = NewBatchEngine(maxCount, timeout, func(engine *BatchEngine[SensorData]) error {
		return save(sdb)
	})
	return sdb, err
}

func (sdb *SensorDataBatch) AddSensorData(data SensorData) {
	sdb.engine.Add(data)
}

func (sdb *SensorDataBatch) Count() int {
	return sdb.engine.Count()
}

func (sdb *SensorDataBatch) Items() []SensorData {
	return sdb.engine.Items()
}

func (sdb *SensorDataBatch) GetKafkaMessages() []kafka.Message {
	messages := make([]kafka.Message, 0, sdb.Count())
	for _, d := range sdb.Items() {
		if d.KafkaMsg.Value != nil {
			messages = append(messages, d.KafkaMsg)
		}
	}
	return messages
}

// AggregatedStats contiene i dati statistici calcolati ogni tot minuti dal Proximity-Fog-Hub
// e inviati tramite kafka all' intermediate-fog-hub per essere memorizzati nel database centrale
type AggregatedStats struct {
	ID            string  `json:"id,omitempty"`
	Timestamp     int64   `json:"timestamp"`
	Region        string  `json:"region,omitempty"`
	Macrozone     string  `json:"macrozone,omitempty"`
	Zone          string  `json:"zone,omitempty"`
	Type          string  `json:"type"`
	Min           float64 `json:"min"`
	Max           float64 `json:"max"`
	Avg           float64 `json:"avg"`
	Sum           float64 `json:"sum,omitempty"`
	Count         int     `json:"count,omitempty"`
	WeightedAvg   float64 `json:"weighted_avg,omitempty"`
	WeightedSum   float64 `json:"weighted_sum,omitempty"`
	WeightedCount float64 `json:"weighted_count,omitempty"`

	KafkaMsg kafka.Message `json:"-"`
}

// CreateAggregatedStatsFromKafka deserializza un messaggio Kafka in AggregatedStats
func CreateAggregatedStatsFromKafka(msg kafka.Message) (AggregatedStats, error) {
	var stats AggregatedStats
	err := json.Unmarshal(msg.Value, &stats)
	stats.KafkaMsg = msg
	return stats, err
}

type AggregatedStatsBatch struct {
	engine *BatchEngine[AggregatedStats]
}

func NewAggregatedStatsBatch(maxCount int, timeout time.Duration, save func(*AggregatedStatsBatch) error) (*AggregatedStatsBatch, error) {
	asb := &AggregatedStatsBatch{}
	var err error
	asb.engine, err = NewBatchEngine(maxCount, timeout, func(engine *BatchEngine[AggregatedStats]) error {
		return save(asb)
	})
	return asb, err
}

func (asb *AggregatedStatsBatch) AddAggregatedStats(stats AggregatedStats) {
	asb.engine.Add(stats)
}

func (asb *AggregatedStatsBatch) Count() int {
	return asb.engine.Count()
}

func (asb *AggregatedStatsBatch) Items() []AggregatedStats {
	return asb.engine.Items()
}

func (asb *AggregatedStatsBatch) GetKafkaMessages() []kafka.Message {
	messages := make([]kafka.Message, 0, asb.Count())
	for _, s := range asb.Items() {
		if s.KafkaMsg.Value != nil {
			messages = append(messages, s.KafkaMsg)
		}
	}
	return messages
}
