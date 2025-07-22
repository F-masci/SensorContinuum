package comunication

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
)

var kafkaReader *kafka.Reader = nil
var kafkaWriter *kafka.Writer = nil

func connect() {
	if kafkaReader != nil && kafkaWriter != nil {
		return // already connected
	}

	// Configure the Kafka reader
	kafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.EdgeHubTopic,
		GroupID: environment.BuildingID,
	})
	logger.Log.Info("Connected (reader) to Kafka topic: ", environment.EdgeHubTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configura il writer Kafka
	kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityDataTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (writing) to Kafka topic: ", environment.ProximityDataTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

}

func PullFilteredData(dataChannel chan structure.SensorData) error {

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
