package comunication

import (
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

var kafkaWriter *kafka.Writer = nil

func connect() {
	if kafkaWriter != nil {
		return // già connesso
	}

	// Configura il writer Kafka
	// Scrive sulla partizione relativa al piano
	kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.EdgeHubTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
}

func SendAggregatedData(data structure.SensorData) error {

	connect()

	// Serializza il dato in JSON
	msgBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Invia il messaggio sulla partizione desiderata
	logger.Log.Debug("Sending message to Kafka topic: ", environment.EdgeHubTopic)
	return kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(environment.EdgeHubTopicPartition),
			Value: msgBytes,
		},
	)

}
