package types

import (
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/segmentio/kafka-go"
)

type HeartbeatMsg struct {
	Timestamp     int64  `json:"timestamp,omitempty"`
	EdgeMacrozone string `json:"macrozone,omitempty"`
	EdgeZone      string `json:"zone,omitempty"`
	HubID         string `json:"hub_id,omitempty"`
}

// CreateHeartbeatMsgFromKafka crea un messaggio di heartbeat da un messaggio Kafka
func CreateHeartbeatMsgFromKafka(msg kafka.Message) (HeartbeatMsg, error) {
	var heartbeatMsg HeartbeatMsg
	err := json.Unmarshal(msg.Value, &heartbeatMsg)
	return heartbeatMsg, err
}

func CreateHeartbeatMsgFromMqtt(msg mqtt.Message) (HeartbeatMsg, error) {
	var heartbeatMsg HeartbeatMsg

	if msg == nil || msg.Payload() == nil || len(msg.Payload()) == 0 || msg.Topic() == "" {
		return HeartbeatMsg{}, nil
	}

	err := json.Unmarshal(msg.Payload(), &heartbeatMsg)
	return heartbeatMsg, err
}
