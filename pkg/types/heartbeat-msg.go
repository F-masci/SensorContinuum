package types

import (
	"encoding/json"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/segmentio/kafka-go"
)

type HeartbeatMsg struct {
	Timestamp     int64  `json:"timestamp,omitempty"`
	EdgeMacrozone string `json:"macrozone,omitempty"`
	EdgeZone      string `json:"zone,omitempty"`
	HubID         string `json:"hub_id,omitempty"`

	KafkaMsg kafka.Message `json:"-"`
	MQTTMsg  mqtt.Message  `json:"-"`
}

// CreateHeartbeatMsgFromKafka crea un messaggio di heartbeat da un messaggio Kafka
func CreateHeartbeatMsgFromKafka(msg kafka.Message) (HeartbeatMsg, error) {
	var heartbeatMsg HeartbeatMsg
	err := json.Unmarshal(msg.Value, &heartbeatMsg)
	heartbeatMsg.KafkaMsg = msg
	return heartbeatMsg, err
}

func CreateHeartbeatMsgFromMqtt(msg mqtt.Message) (HeartbeatMsg, error) {
	var heartbeatMsg HeartbeatMsg

	if msg == nil || msg.Payload() == nil || len(msg.Payload()) == 0 || msg.Topic() == "" {
		return HeartbeatMsg{}, nil
	}

	err := json.Unmarshal(msg.Payload(), &heartbeatMsg)
	heartbeatMsg.MQTTMsg = msg
	return heartbeatMsg, err
}

type HeartbeatMsgBatch struct {
	engine *BatchEngine[HeartbeatMsg]
}

func NewHeartbeatMsgBatch(maxCount int, timeout time.Duration, save func(*HeartbeatMsgBatch) error) (*HeartbeatMsgBatch, error) {
	hbb := &HeartbeatMsgBatch{}
	var err error
	hbb.engine, err = NewBatchEngine(maxCount, timeout, func(engine *BatchEngine[HeartbeatMsg]) error {
		return save(hbb)
	})
	return hbb, err
}

func (hbb *HeartbeatMsgBatch) Add(msg HeartbeatMsg) {
	hbb.engine.Add(msg)
}

func (hbb *HeartbeatMsgBatch) Count() int {
	return hbb.engine.Count()
}

func (hbb *HeartbeatMsgBatch) Items() []HeartbeatMsg {
	return hbb.engine.Items()
}

func (hbb *HeartbeatMsgBatch) GetKafkaMessages() []kafka.Message {
	messages := make([]kafka.Message, 0, hbb.Count())
	for _, d := range hbb.Items() {
		if d.KafkaMsg.Value != nil {
			messages = append(messages, d.KafkaMsg)
		}
	}
	return messages
}
