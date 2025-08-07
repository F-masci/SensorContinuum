package structure

import (
	"encoding/json"
	"github.com/segmentio/kafka-go"
)

type ConfigurationMsg struct {
	BuildingID string `json:"building_id"`
	MsgType    string `json:"msg_type"`
	Timestamp  int64  `json:"timestamp"`
}

func CreateConfigurationMsgFromKafka(msg kafka.Message) (ConfigurationMsg, error) {
	var confMsg ConfigurationMsg
	err := json.Unmarshal(msg.Value, &confMsg)
	return confMsg, err
}
