package comunication

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"time"
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

	// Creiamo un contesto con un timeout di 5 secondi.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Log.Debug("Sending message to Kafka topic", "topic", environment.ProximityDataTopic)
	// Usiamo il contesto nella chiamata di scrittura.
	return kafkaWriter.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(environment.ProximityDataTopicPartition),
			Value: msgBytes,
		},
	)

}
