package aggregation

import (
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"context"
	"time"
)

// PerformAggregationAndSend è la funzione che viene eseguita periodicamente
func PerformAggregationAndSend() {
	logger.Log.Info("execution of periodic aggregation started")
	ctx := context.Background()

	// 1. Esegui la query per ottenere le statistiche
	stats, err := storage.GetValueToSend(ctx)
	if err != nil {
		logger.Log.Error("failed to calculate periodic statistics, error: ", err)
		return
	}

	if len(stats) == 0 {
		logger.Log.Info("No data to send, skipping aggregation")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// 2. Invia ogni statistica a Kafka
	for _, stat := range stats {
		// Arricchiamo la statistica con dati contestuali
		stat.Timestamp = now
		stat.Macrozone = environment.EdgeMacrozone

		logger.Log.Info("Statistics calculated for the type: ", stat.Type, " - avg:", stat.Avg)

		if err := comunication.SendData(stat); err != nil {
			logger.Log.Error("Failure to send statistics to Kafka, type", stat.Type, " - error: ", err)
			// Non ci fermiamo, proviamo a inviare le altre
			continue
		}
		logger.Log.Info("Statistics successfully sent to Kafka for type: ", stat.Type)
	}
}
