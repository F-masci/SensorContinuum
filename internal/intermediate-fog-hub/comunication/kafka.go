package comunication

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"

	"github.com/segmentio/kafka-go"
)

// kafkaRealTimeDataReader è il lettore Kafka per i dati in tempo reale.
var kafkaRealTimeDataReader *kafka.Reader = nil

// kafkaStatisticsDataReader è il lettore Kafka per i dati statistici aggregati.
var kafkaStatisticsDataReader *kafka.Reader = nil

// kafkaConfigurationReader è il lettore Kafka per i messaggi di configurazione.
var kafkaConfigurationReader *kafka.Reader = nil

// kafkaHeartbeatReader è il lettore Kafka per i messaggi di heartbeat.
var kafkaHeartbeatReader *kafka.Reader = nil

// connectRealTimeData si connette a Kafka per leggere i dati in tempo reale.
func connectRealTimeData() {

	// Se la connessione è già stabilita, non fare nulla
	if kafkaRealTimeDataReader != nil {
		return // already connected
	}

	logger.Log.Debug("Connecting to Kafka topic: ", environment.ProximityDataTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configura il lettore Kafka per i dati in tempo reale
	kafkaRealTimeDataReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.ProximityDataTopic,
		GroupID: environment.KafkaGroupId,
	})
	logger.Log.Info("Connected to Kafka topic: ", environment.ProximityDataTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)
}

// connectProximityConfiguration si connette a Kafka per leggere i messaggi di configurazione.
func connectProximityConfiguration() {

	// Se la connessione è già stabilita, non fare nulla
	if kafkaConfigurationReader != nil {
		return // already connected
	}

	logger.Log.Debug("Connecting to Kafka topic: ", environment.ProximityConfigurationTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configure the Kafka reader
	kafkaConfigurationReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.ProximityConfigurationTopic,
		GroupID: environment.KafkaGroupId,
	})
	logger.Log.Info("Connected to Kafka topic: ", environment.ProximityConfigurationTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)
}

// connectProximityHeartbeat si connette a Kafka per leggere i messaggi di heartbeat.
func connectProximityHeartbeat() {

	// Se la connessione è già stabilita, non fare nulla
	if kafkaHeartbeatReader != nil {
		return
	}

	logger.Log.Debug("Connecting to Kafka topic: ", environment.ProximityHeartbeatTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configure the Kafka reader
	kafkaHeartbeatReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.ProximityHeartbeatTopic,
		GroupID: environment.KafkaGroupId,
	})
	logger.Log.Info("Connected to Kafka topic: ", environment.ProximityHeartbeatTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)
}

// connectStatisticsData si connette a Kafka per leggere i dati statistici aggregati.
func connectStatisticsData() {

	// Se la connessione è già stabilita, non fare nulla
	if kafkaStatisticsDataReader != nil {
		return // already connected
	}

	logger.Log.Debug("Connecting to Kafka topic: ", environment.AggregatedStatsTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configure the Kafka reader
	kafkaStatisticsDataReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.AggregatedStatsTopic,
		GroupID: environment.KafkaGroupId,
	})
	logger.Log.Info("Connected to Kafka topic: ", environment.AggregatedStatsTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)
}

// PullRealTimeData si occupa di leggere i dati dei sensori in tempo reale.
func PullRealTimeData(dataChannel chan types.SensorData) error {

	// Connessione a Kafka se non è già stabilita
	connectRealTimeData()

	ctx := context.Background()
	for {
		// Legge il messaggio dal topic Kafka
		m, err := kafkaRealTimeDataReader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))

		// Converte il messaggio in un oggetto SensorData
		var data types.SensorData
		data, err = types.CreateSensorDataFromKafka(m)
		if err != nil {
			logger.Log.Error("Error unmarshalling Sensor Data: ", err)
			continue
		}

		// Prova a inviare il dato al canale senza bloccare
		// Se il canale è pieno, scarta il dato e logga un avviso
		select {
		case dataChannel <- data:
			// Inviato con successo
			logger.Log.Debug("Sensor data sent to channel: ", data)
		default:
			// Canale pieno, logghiamo un avviso e scartiamo il dato
			logger.Log.Warn("Data channel is full, discarding sensor data: ", data)
		}
	}
}

// PullStatisticsData si occupa di leggere i dati statistici aggregati.
func PullStatisticsData(statsChannel chan types.AggregatedStats) error {

	// Connessione a Kafka se non è già stabilita
	connectStatisticsData()

	ctx := context.Background()
	for {
		// Legge il messaggio dal topic Kafka
		m, err := kafkaStatisticsDataReader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		if err != nil {
			logger.Log.Error("Error reading message: ", err.Error())
			return err
		}
		logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))

		// Converte il messaggio in un oggetto AggregatedStats
		var stats types.AggregatedStats
		stats, err = types.CreateAggregatedStatsFromKafka(m)
		if err != nil {
			logger.Log.Error("Error unmarshalling Aggregated Stats", "error", err)
			continue
		}

		// Invia il dato al canale
		select {
		case statsChannel <- stats:
			// Inviato con successo
			logger.Log.Debug("Aggregated stats sent to channel: ", stats)
		default:
			// Canale pieno, logghiamo un avviso e scartiamo il dato
			logger.Log.Warn("Stats channel is full, discarding aggregated stats: ", stats)
		}
	}
}

// PullConfigurationMessage si occupa di leggere i messaggi di configurazione.
func PullConfigurationMessage(msgChannel chan types.ConfigurationMsg) error {

	// Connessione a Kafka se non è già stabilita
	connectProximityConfiguration()

	ctx := context.Background()
	for {
		// Legge il messaggio dal topic Kafka
		m, err := kafkaConfigurationReader.ReadMessage(ctx)
		if err != nil {
			logger.Log.Error("Error reading message: ", err.Error())
			return err
		}
		logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))

		// Converte il messaggio in un oggetto ConfigurationMsg
		var msg types.ConfigurationMsg
		msg, err = types.CreateConfigurationMsgFromKafka(m)
		if err != nil {
			logger.Log.Error("Error unmarshalling Sensor Data: ", err.Error())
			continue
		}

		// Invia il messaggio al canale
		select {
		case msgChannel <- msg:
			// Inviato con successo
			logger.Log.Debug("Configuration message sent to channel: ", msg)
		default:
			// Canale pieno, logghiamo un avviso e scartiamo il messaggio
			logger.Log.Warn("Configuration message channel is full, discarding message: ", msg)
		}
	}
}

// PullHeartbeatMessage si occupa di leggere i messaggi di heartbeat.
func PullHeartbeatMessage(heartbeatChannel chan types.HeartbeatMsg) error {

	// Connessione a Kafka se non è già stabilita
	connectProximityHeartbeat()

	ctx := context.Background()
	for {
		// Legge il messaggio dal topic Kafka
		m, err := kafkaHeartbeatReader.ReadMessage(ctx)
		if err != nil {
			logger.Log.Error("Error reading message: ", err.Error())
			return err
		}
		logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))

		// Converte il messaggio in un oggetto HeartbeatMsg
		var heartbeatMsg types.HeartbeatMsg
		heartbeatMsg, err = types.CreateHeartbeatMsgFromKafka(m)
		if err != nil {
			logger.Log.Error("Error unmarshalling Heartbeat Message: ", err.Error())
			continue
		}

		// Invia il messaggio al canale
		select {
		case heartbeatChannel <- heartbeatMsg:
			// Inviato con successo
			logger.Log.Debug("Heartbeat message sent to channel: ", heartbeatMsg)
		default:
			// Canale pieno, logghiamo un avviso e scartiamo il messaggio
			logger.Log.Warn("Heartbeat message channel is full, discarding message: ", heartbeatMsg)
		}
	}
}
