package comunication

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/internal/proximity-fog-hub/storage" // --- NUOVA IMPORTAZIONE ---
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"time"
)

var kafkaWriter *kafka.Writer = nil
var statsKafkaWriter *kafka.Writer = nil // --- NUOVO WRITER PER LE STATISTICHE ---

func connect() {
	if kafkaWriter != nil {
		return
	}

	kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityDataTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (write) to Kafka topic for real-time data, topic: ", environment.ProximityDataTopic)

	// --- NUOVA LOGICA: Connessione per il topic delle statistiche ---
	statsKafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.KafkaAggregatedStatsTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (write) to Kafka topic for real-time data, topic:", environment.KafkaAggregatedStatsTopic)
}

func SendAggregatedData(data structure.SensorData) error {
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
func SendData(stats storage.AggregatedStats) error {
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
