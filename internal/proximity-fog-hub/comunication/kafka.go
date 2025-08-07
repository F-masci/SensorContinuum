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

// writer per le misurazioni in tempo reale
var kafkaWriter *kafka.Writer = nil

// writer per le statistiche aggregate
var statsKafkaWriter *kafka.Writer = nil

func connect() {
	if kafkaWriter != nil {
		return
	}

	// connessione per il topic delle misurazioni in tempo reale
	kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityDataTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (write) to Kafka topic for real-time data, topic: ", environment.ProximityDataTopic)

	// Connessione per il topic delle statistiche
	statsKafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.KafkaAggregatedStatsTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (write) to Kafka topic for real-time data, topic:", environment.KafkaAggregatedStatsTopic)
}

func SendRealTimeData(data structure.SensorData) error {
	connect()
	msgBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return kafkaWriter.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(environment.ProximityDataTopicPartition),
			Value: msgBytes,
		},
	)
}

// SendData invia le statistiche aggregate al topic Kafka dedicato
func SendData(stats structure.AggregatedStats) error {
	connect()

	msgBytes, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return statsKafkaWriter.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(environment.BuildingID), // Partizioniamo per edificio
			Value: msgBytes,
		},
	)
}
