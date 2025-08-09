package comunication

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"github.com/segmentio/kafka-go"
)

var kafkaDataReader *kafka.Reader = nil
var kafkaConfigurationReader *kafka.Reader = nil

func connectProximityData() {
	if kafkaDataReader != nil {
		return // already connected
	}

	logger.Log.Debug("Connecting to Kafka topic: ", environment.ProximityDataTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configure the Kafka reader
	kafkaDataReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.ProximityDataTopic,
		GroupID: "intermediate-fog-hub", // Il fog gestisce una singola regione
	})
	logger.Log.Info("Connected to Kafka topic: ", environment.ProximityDataTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)
}

func connectProximityConfiguration() {
	if kafkaConfigurationReader != nil {
		return // already connected
	}

	logger.Log.Debug("Connecting to Kafka topic: ", environment.ProximityConfigurationTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configure the Kafka reader
	kafkaConfigurationReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.ProximityConfigurationTopic,
		GroupID: "intermediate-fog-hub", // Il fog gestisce una singola regione
	})
	logger.Log.Info("Connected to Kafka topic: ", environment.ProximityConfigurationTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)
}

// PullRealTimeData si occupa di leggere i dati dei sensori in tempo reale.
func PullRealTimeData(dataChannel chan types.SensorData) error {

	connectProximityData()

	ctx := context.Background()
	for {
		m, err := kafkaDataReader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))
		var data types.SensorData
		data, err = types.CreateSensorDataFromKafka(m)
		if err != nil {
			logger.Log.Error("Errore nel deserializzare SensorData", "error", err)
			continue
		}
		dataChannel <- data
	}
}

// PullStatisticsData si occupa di leggere i dati statistici aggregati.
func PullStatisticsData(statsChannel chan types.AggregatedStats) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.KafkaAggregatedStatsTopic,
		GroupID: "intermediate-fog-hub", // Il fog gestisce una singola regione
	})
	defer reader.Close()
	logger.Log.Info("Consumatore avviato per statistiche", "topic", environment.KafkaAggregatedStatsTopic)

	ctx := context.Background()
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		stats, err := structure.CreateAggregatedStatsFromKafka(m)
		if err != nil {
			logger.Log.Error("Errore nel deserializzare AggregatedStats", "error", err)
			continue
		}
		statsChannel <- stats
	}
}

func PullConfigurationMessage(msgChannel chan types.ConfigurationMsg) error {

	connectProximityConfiguration()

	ctx := context.Background()
	for {
		m, err := kafkaConfigurationReader.ReadMessage(ctx)
		if err != nil {
			logger.Log.Error("Error reading message: ", err.Error())
			return err
		}
		logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))
		var msg types.ConfigurationMsg
		msg, err = types.CreateConfigurationMsgFromKafka(m)
		if err != nil {
			logger.Log.Error("Error unmarshalling Sensor Data: ", err.Error())
			continue
		}
		msgChannel <- msg
	}

}
