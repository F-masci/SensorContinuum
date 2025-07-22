package comunication

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
	"github.com/segmentio/kafka-go"
)

var kafkaReader *kafka.Reader = nil

func connect() {
	if kafkaReader != nil {
		return // already connected
	}

	// Configure the Kafka reader
	kafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.ProximityDataTopic,
		GroupID: environment.BuildingID,
	})
	logger.Log.Info("Connected to Kafka topic: ", environment.ProximityDataTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)
}

func PullAggregatedData(dataChannel chan structure.SensorData) error {

	connect()

	ctx := context.Background()
	for {
		m, err := kafkaReader.ReadMessage(ctx)
		logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))
		if err != nil {
			logger.Log.Error("Error reading message: ", err.Error())
			return err
		}
		var data structure.SensorData
		data, err = structure.CreateSensorDataFromKafka(m)
		if err != nil {
			logger.Log.Error("Error unmarshalling Sensor Data: ", err.Error())
			continue
		}
		dataChannel <- data
	}

}
