package comunication

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"github.com/segmentio/kafka-go"
)

var kafkaRealTimeDataReader *kafka.Reader = nil
var kafkaConfigurationReader *kafka.Reader = nil
var kafkaStatisticsDataReader *kafka.Reader = nil

func connectRealTimeData() {
	if kafkaRealTimeDataReader != nil {
		return // already connected
	}

	logger.Log.Debug("Connecting to Kafka topic: ", environment.ProximityDataTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configure the Kafka reader
	kafkaRealTimeDataReader = kafka.NewReader(kafka.ReaderConfig{
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

func connectStatisticsData() {
	if kafkaStatisticsDataReader != nil {
		return // already connected
	}

	logger.Log.Debug("Connecting to Kafka topic: ", environment.KafkaAggregatedStatsTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configure the Kafka reader
	kafkaStatisticsDataReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.KafkaAggregatedStatsTopic,
		GroupID: "intermediate-fog-hub", // Il fog gestisce una singola regione
	})
	logger.Log.Info("Connected to Kafka topic: ", environment.KafkaAggregatedStatsTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)
}

// PullRealTimeData si occupa di leggere i dati dei sensori in tempo reale.
func PullRealTimeData(dataChannel chan types.SensorData) error {

	connectRealTimeData()

	ctx := context.Background()
	for {
		m, err := kafkaRealTimeDataReader.ReadMessage(ctx)
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

	connectStatisticsData()

	ctx := context.Background()
	for {
		m, err := kafkaStatisticsDataReader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		if err != nil {
			logger.Log.Error("Error reading message: ", err.Error())
			return err
		}
		logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))
		var stats types.AggregatedStats
		stats, err = types.CreateAggregatedStatsFromKafka(m)
		if err != nil {
			logger.Log.Error("Error unmarshalling Aggregated Stats", "error", err)
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
