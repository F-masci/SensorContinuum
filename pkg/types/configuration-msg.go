package types

import (
	"encoding/json"
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
}

func CreateConfigurationMsgFromKafka(msg kafka.Message) (ConfigurationMsg, error) {
	var confMsg ConfigurationMsg
	err := json.Unmarshal(msg.Value, &confMsg)
	return confMsg, err
}

func CreateConfigurationMsgFromMqtt(msg mqtt.Message) (ConfigurationMsg, error) {
	var edgeConfMsg ConfigurationMsg

	if msg == nil || msg.Payload() == nil || len(msg.Payload()) == 0 || msg.Topic() == "" {
		return ConfigurationMsg{}, nil
	}

	err := json.Unmarshal(msg.Payload(), &edgeConfMsg)
	return edgeConfMsg, err
}
