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

// realtimeKafkaWriter per le misurazioni in tempo reale
var realtimeKafkaWriter *kafka.Writer = nil

// statsKafkaWriter per le statistiche aggregate
var statsKafkaWriter *kafka.Writer = nil

// configurationKafkaWriter per i messaggi di configurazione
var configurationKafkaWriter *kafka.Writer = nil

// heartbeatKafkaWriter per i messaggi di heartbeat
var heartbeatKafkaWriter *kafka.Writer = nil

func connect() {
	if realtimeKafkaWriter != nil && statsKafkaWriter != nil && configurationKafkaWriter != nil {
		return
	}

	// connessione per il topic delle misurazioni in tempo reale
	realtimeKafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityRealtimeDataTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (write) to Kafka topic for real-time data, topic: ", environment.ProximityRealtimeDataTopic)

	// Connessione per il topic delle statistiche
	statsKafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityAggregatedStatsTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (write) to Kafka topic for stats data, topic: ", environment.ProximityAggregatedStatsTopic)

	// Connessione per il topic della configurazione ---
	configurationKafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityConfigurationTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (write) to Kafka topic for configuration data, topic: ", environment.ProximityConfigurationTopic)

	//Connessione per il topic dei messaggi di heartbeat ---
	heartbeatKafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityHeartbeatTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
}

func SendRealTimeData(data types.SensorData) error {
	connect()
	msgBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return realtimeKafkaWriter.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(environment.ProximityRealtimeDataTopicPartition),
			Value: msgBytes,
		},
	)
}

// SendAggregatedData invia le statistiche aggregate al topic Kafka dedicato
func SendAggregatedData(stats types.AggregatedStats) error {
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

func SendHeartbeatMessage(msg types.HeartbeatMsg) error {
	connect()

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := msg.EdgeMacrozone + "-" + msg.EdgeZone + "-" + msg.HubID
	return heartbeatKafkaWriter.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(key),
			Value: msgBytes,
		},
	)

}
