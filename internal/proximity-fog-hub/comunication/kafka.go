package comunication

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"time"
)

// writer per le misurazioni in tempo reale
var kafkaWriter *kafka.Writer = nil

// writer per le statistiche aggregate
var statsKafkaWriter *kafka.Writer = nil

// writer per i messaggi di configurazione
var configurationKafkaWriter *kafka.Writer = nil

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
	logger.Log.Info("Connected (write) to Kafka topic for stats data, topic: ", environment.KafkaAggregatedStatsTopic)

	// --- NUOVA LOGICA: Connessione per il topic della configurazione ---
	configurationKafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityConfigurationTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (write) to Kafka topic for configuration data, topic: ", environment.ProximityConfigurationTopic)
}

func SendRealTimeData(data types.SensorData) error {
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
func SendData(stats types.AggregatedStats) error {
	connect()

	msgBytes, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return statsKafkaWriter.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(environment.EdgeMacrozone), // Partizioniamo per edificio
			Value: msgBytes,
		},
	)
}

func SendRegistrationMessage() error {

	msg := types.ConfigurationMsg{
		MsgType:       types.NewProximityMsgType,
		EdgeMacrozone: environment.EdgeMacrozone,
		Timestamp:     time.Now().Unix(),
		HubID:         environment.HubID,
		Service:       types.ProximityHubService,
	}

	return SendConfigurationMessage(msg)
}

func SendConfigurationMessage(msg types.ConfigurationMsg) error {
	connect()

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return configurationKafkaWriter.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(environment.EdgeMacrozone),
			Value: msgBytes,
		},
	)
}
