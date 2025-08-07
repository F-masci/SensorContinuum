package comunication

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
	"github.com/segmentio/kafka-go"
)

// PullRealTimeData si occupa di leggere i dati dei sensori in tempo reale.
func PullRealTimeData(dataChannel chan<- structure.SensorData) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.ProximityDataTopic,
		GroupID: environment.BuildingID + "-realtime", // GroupID univoco
	})
	defer reader.Close()
	logger.Log.Info("Consumatore avviato per dati real-time", "topic", environment.ProximityDataTopic)

	ctx := context.Background()
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		data, err := structure.CreateSensorDataFromKafka(m)
		if err != nil {
			logger.Log.Error("Errore nel deserializzare SensorData", "error", err)
			continue
		}
		dataChannel <- data
	}
}

// PullStatisticsData si occupa di leggere i dati statistici aggregati.
func PullStatisticsData(statsChannel chan<- structure.AggregatedStats) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.KafkaAggregatedStatsTopic,
		GroupID: environment.BuildingID + "-statistics", // GroupID univoco
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
