package comunication

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
)

var kafkaWriter *kafka.Writer = nil

func connect() {
	if kafkaWriter != nil {
		return // already connected
	}

	// Configura il writer Kafka
	kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityDataTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (writing) to Kafka topic: ", environment.ProximityDataTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

}

func SendAggregatedData(data structure.SensorData) error {

	connect()

	// Serializza il dato in JSON
	msgBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Invia il messaggio sulla partizione desiderata
	logger.Log.Debug("Sending message to Kafka topic: ", environment.ProximityDataTopic)
	return kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(environment.ProximityDataTopicPartition),
			Value: msgBytes,
		},
	)

}
