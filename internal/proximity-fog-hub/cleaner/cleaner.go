package cleaner

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"context"
	"time"
)

// Run avvia il processo di pulizia della tabella outbox, è molto simile a quella del dispatcher.
// Essa usa infatti un ticker e un ctx per uno spegnimento pulito
// Questa funzione viene eseguita in una goroutine separata e si occupa di:
//  1. Attivarsi periodicamente (ogni 5 minuti).
//  2. Cancellare i messaggi dalla tabella 'aggregated_stats_cache' che sono in stato 'sent'
//     e sono più vecchi di una soglia definita (es. 1 ora).
func Run(ctx context.Context) {
	logger.Log.Info("Starting Outbox Cleaner...")
	ticker := time.NewTicker(environment.CleanerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Stopping Outbox Cleaner...")
			return
		case <-ticker.C:
			logger.Log.Info("Outbox Cleaner running...")
			CleanupSentMessages(ctx)
		}
	}
}

// CleanupSentMessages si occupa della logica effettiva di cancellazione.
func CleanupSentMessages(ctx context.Context) {
	// Pulisce i messaggi dei sensori
	CleanupRawSentMessages(ctx)
	// Pulisce i messaggi aggregati
	CleanupAggregatedSentMessages(ctx)
}

// CleanupRawSentMessages si occupa della logica di cancellazione dei messaggi dei sensori.
func CleanupRawSentMessages(ctx context.Context) {
	deletedCount, err := storage.DeleteSensorData(ctx, environment.SentMessageMaxAge)
	if err != nil {
		logger.Log.Error("Error cleaning up sent sensor data outbox messages: ", err)
		return
	}
	if deletedCount > 0 {
		logger.Log.Info("Outbox Cleaner successfully deleted ", deletedCount, " sent sensor data messages.")
	} else {
		logger.Log.Info("Outbox Cleaner found no old sent sensor data messages to delete.")
	}
}

// CleanupAggregatedSentMessages si occupa della logica di cancellazione dei messaggi aggregati.
func CleanupAggregatedSentMessages(ctx context.Context) {
	deletedCount, err := storage.DeleteAggregatedStats(ctx, environment.SentMessageMaxAge)
	if err != nil {
		logger.Log.Error("Error cleaning up sent aggregated stats outbox messages: ", err)
		return
	}
	if deletedCount > 0 {
		logger.Log.Info("Outbox Cleaner successfully deleted ", deletedCount, " sent aggregated stats messages.")
	} else {
		logger.Log.Info("Outbox Cleaner found no old sent aggregated stats messages to delete.")
	}
}
