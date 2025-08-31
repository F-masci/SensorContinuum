package types

import (
	"encoding/json"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/segmentio/kafka-go"
)

type MsgType string

// Tipi di messaggi di configurazione
const (
	NewProximityMsgType MsgType = "new_proximity"
	NewEdgeMsgType      MsgType = "new_edge"
	NewSensorMsgType    MsgType = "new_sensor"
)

type ConfigurationMsg struct {
	MsgType         MsgType `json:"msg_type,omitempty"`
	Service         Service `json:"service,omitempty"`
	Timestamp       int64   `json:"timestamp,omitempty"`
	EdgeMacrozone   string  `json:"macrozone,omitempty"`
	EdgeZone        string  `json:"zone,omitempty"`
	HubID           string  `json:"hub_id,omitempty"`
	SensorID        string  `json:"sensor_id,omitempty"`
	SensorLocation  string  `json:"sensor_location,omitempty"`
	SensorType      string  `json:"sensor_type,omitempty"`
	SensorReference string  `json:"sensor_reference,omitempty"`

	KafkaMsg kafka.Message `json:"-"`
	MQTTMsg  mqtt.Message  `json:"-"`
}

func CreateConfigurationMsgFromKafka(msg kafka.Message) (ConfigurationMsg, error) {
	var confMsg ConfigurationMsg
	err := json.Unmarshal(msg.Value, &confMsg)
	confMsg.KafkaMsg = msg
	return confMsg, err
}

func CreateConfigurationMsgFromMqtt(msg mqtt.Message) (ConfigurationMsg, error) {
	var edgeConfMsg ConfigurationMsg

	if msg == nil || msg.Payload() == nil || len(msg.Payload()) == 0 || msg.Topic() == "" {
		return ConfigurationMsg{}, nil
	}

	err := json.Unmarshal(msg.Payload(), &edgeConfMsg)
	edgeConfMsg.MQTTMsg = msg
	return edgeConfMsg, err
}

type ConfigurationMsgBatch struct {
	engine *BatchEngine[ConfigurationMsg]
}

func NewConfigurationMsgBatch(maxCount int, timeout time.Duration, save func(*ConfigurationMsgBatch) error) (*ConfigurationMsgBatch, error) {
	cmb := &ConfigurationMsgBatch{}
	var err error
	cmb.engine, err = NewBatchEngine(maxCount, timeout, func(engine *BatchEngine[ConfigurationMsg]) error {
		return save(cmb)
	})
	return cmb, err
}

func (cmb *ConfigurationMsgBatch) Add(msg ConfigurationMsg) {
	cmb.engine.Add(msg)
}

func (cmb *ConfigurationMsgBatch) Count() int {
	return cmb.engine.Count()
}

func (cmb *ConfigurationMsgBatch) Items() []ConfigurationMsg {
	return cmb.engine.Items()
}

func (cmb *ConfigurationMsgBatch) GetKafkaMessages() []kafka.Message {
	messages := make([]kafka.Message, 0, cmb.Count())
	for _, c := range cmb.Items() {
		if c.KafkaMsg.Value != nil {
			messages = append(messages, c.KafkaMsg)
		}
	}
	return messages
}
