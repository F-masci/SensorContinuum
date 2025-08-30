package dispatcher

import (
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"context"
	"time"
)

// Run avvia il processo del dispatcher dell'outbox.
// Questa funzione viene eseguita in una goroutine separata e si occupa di:
// 1. Controllare periodicamente la tabella 'aggregated_stats_cache' per messaggi in stato 'pending'.
// 2. Inviare questi messaggi a Kafka.
// 3. Aggiornare lo stato dei messaggi a 'sent' solo dopo un invio andato a buon fine.
func Run(ctx context.Context) {

	// Avvio del ticker per il polling periodico
	outboxTicker := time.NewTicker(environment.OutboxPollInterval)
	logger.Log.Info("Outbox ticker started, sending data every ", environment.OutboxPollInterval, " minutes from now.")
	defer outboxTicker.Stop()

	for {
		select {
		// Permette uno spegnimento pulito quando il contesto viene annullato
		case <-ctx.Done():
			logger.Log.Info("Stopping Outbox Dispatcher...")
			return
		case <-outboxTicker.C:
			logger.Log.Info("Outbox Dispatcher checking for pending messages...")
			ProcessPendingMessages(ctx)
		}
	}
}

// ProcessPendingMessages recupera e processa i messaggi pendenti dalla tabella outbox.
func ProcessPendingMessages(ctx context.Context) {
	// Processa i messaggi di dati grezzi
	ProcessRawPendingMessages(ctx)
	// Processa i messaggi di statistiche aggregate
	ProcessAggregatedPendingMessages(ctx)
}

// ProcessRawPendingMessages recupera e processa i messaggi pendenti dalla tabella outbox.
func ProcessRawPendingMessages(ctx context.Context) {

	var attempts int = 0
	nMessages := environment.OutboxBatchSize

	for nMessages == environment.OutboxBatchSize {

		// 1. Recupera i messaggi pendenti dal database
		messages, err := storage.GetPendingSensorData(ctx, environment.OutboxBatchSize)
		if err != nil {
			logger.Log.Error("Error getting pending sensor data outbox messages: ", err)
			return
		}

		if len(messages) == 0 {
			logger.Log.Info("No pending sensor data messages found in outbox.")
			return
		}

		logger.Log.Info("Found ", len(messages), " pending sensor data messages to dispatch.")

		// 2. Invia i messaggi a Kafka
		if err := comunication.SendRealTimeData(messages); err != nil {
			logger.Log.Error("Failed to send sensor data outbox messages to Kafka: ", err)
			attempts++
			if attempts >= environment.OutboxMaxAttempts {
				logger.Log.Error("Max attempts reached for sending sensor data outbox messages. Will retry in next cycle.")
				return
			}
			// Se l'invio fallisce, non facciamo nulla. Il messaggio rimane 'pending'
			// e verrà ritentato al prossimo ciclo.
			continue
		}

		attempts = 0 // reset degli tentativi dopo un invio riuscito

		// 3. Se l'invio ha successo, aggiorna lo stato nel database
		if err := storage.UpdateSensorData(ctx, messages, "sent"); err != nil {
			logger.Log.Error("Failed to update sensor data outbox message status to 'sent': ", err)
			attempts++
			if attempts >= environment.OutboxMaxAttempts {
				logger.Log.Error("Max attempts reached for updating sensor data outbox message status. Will retry in next cycle.")
				return
			}
			// Questo è uno scenario critico: il messaggio è stato inviato ma non siamo riusciti
			// a marcare come tale. Questo causerà un reinvio, che dovrà essere gestito
			// dal consumatore (idempotenza necessaria).
			continue
		}

		attempts = 0 // reset degli tentativi dopo un invio riuscito

		logger.Log.Info("Successfully dispatched sensor data outbox messages.")
		nMessages = len(messages)
	}

}

// ProcessAggregatedPendingMessages recupera e processa i messaggi pendenti dalla tabella outbox.
func ProcessAggregatedPendingMessages(ctx context.Context) {

	var attempts int = 0
	nMessages := environment.OutboxBatchSize

	for nMessages == environment.OutboxBatchSize {

		// 1. Recupera i messaggi pendenti dal database
		messages, err := storage.GetPendingAggregatedStats(ctx, environment.OutboxBatchSize)
		if err != nil {
			logger.Log.Error("Error getting pending aggregated stats outbox messages: ", err)
			return
		}

		if len(messages) == 0 {
			logger.Log.Info("No pending aggregated stats messages found in outbox.")
			return
		}

		logger.Log.Info("Found ", len(messages), " pending aggregated stats messages to dispatch.")

		// 2. Invia i messaggi a Kafka
		if err := comunication.SendAggregatedData(messages); err != nil {
			logger.Log.Error("Failed to send aggregated stats outbox messages to Kafka: ", err)
			attempts++
			if attempts >= environment.OutboxMaxAttempts {
				logger.Log.Error("Max attempts reached for sending aggregated stats outbox messages. Will retry in next cycle.")
				return
			}
			continue
		}

		attempts = 0 // reset tentativi

		// 3. Se l'invio ha successo, aggiorna lo stato nel database
		if err := storage.UpdateAggregatedStats(ctx, messages, "sent"); err != nil {
			logger.Log.Error("Failed to update aggregated stats outbox message status to 'sent': ", err)
			attempts++
			if attempts >= environment.OutboxMaxAttempts {
				logger.Log.Error("Max attempts reached for updating aggregated stats outbox message status. Will retry in next cycle.")
				return
			}
			continue
		}

		attempts = 0 // reset tentativi

		logger.Log.Info("Successfully dispatched aggregated stats outbox messages.")
		nMessages = len(messages)
	}
}
