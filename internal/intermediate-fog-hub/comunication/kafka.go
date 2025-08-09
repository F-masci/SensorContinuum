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

func PullAggregatedData(dataChannel chan types.SensorData) error {

	connectProximityData()

	ctx := context.Background()
	for {
		m, err := kafkaDataReader.ReadMessage(ctx)
		if err != nil {
			logger.Log.Error("Error reading message: ", err.Error())
			return err
		}
		logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))
		var data types.SensorData
		data, err = types.CreateSensorDataFromKafka(m)
		if err != nil {
			logger.Log.Error("Error unmarshalling Sensor Data: ", err.Error())
			continue
		}
		dataChannel <- data
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
