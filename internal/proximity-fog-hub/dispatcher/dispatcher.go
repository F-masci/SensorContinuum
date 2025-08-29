package dispatcher

import (
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"context"
	"time"
)

const (
	// outboxPollInterval definisce ogni quanto il dispatcher controlla la tabella outbox.
	outboxPollInterval = 2 * time.Minute
	// outboxBatchSize definisce quanti messaggi il dispatcher tenta di inviare in ogni ciclo.
	outboxBatchSize = 50
)

// Run avvia il processo del dispatcher dell'outbox.
// Questa funzione viene eseguita in una goroutine separata e si occupa di:
// 1. Controllare periodicamente la tabella 'aggregated_stats_outbox' per messaggi in stato 'pending'.
// 2. Inviare questi messaggi a Kafka.
// 3. Aggiornare lo stato dei messaggi a 'sent' solo dopo un invio andato a buon fine.
func Run(ctx context.Context) {
	logger.Log.Info("Starting Outbox Dispatcher...")
	ticker := time.NewTicker(outboxPollInterval)
	defer ticker.Stop()

	for {
		select {
		// Permette uno spegnimento pulito quando il contesto viene annullato
		case <-ctx.Done():
			logger.Log.Info("Stopping Outbox Dispatcher...")
			return
		case <-ticker.C:
			logger.Log.Info("Outbox Dispatcher checking for pending messages...")
			processPendingMessages(ctx)
		}
	}
}

// processPendingMessages recupera e processa i messaggi pendenti dalla tabella outbox.
func processPendingMessages(ctx context.Context) {
	// 1. Recupera i messaggi pendenti dal database
	messages, err := storage.GetPendingOutboxMessages(ctx, outboxBatchSize)
	if err != nil {
		logger.Log.Error("Error getting pending outbox messages: ", err)
		return
	}

	if len(messages) == 0 {
		logger.Log.Info("No pending messages found in outbox.")
		return
	}

	logger.Log.Info("Found ", len(messages), " pending messages to dispatch.")

	// 2. Itera sui messaggi e tenta di inviarli
	for _, msg := range messages {
		// Aggiungiamo l'ID univoco dal DB al payload, sarà fondamentale per l'idempotenza
		msg.Payload.ID = msg.ID.String()

		// 3. Invia il messaggio a Kafka
		if err := comunication.SendAggregatedData(msg.Payload); err != nil {
			logger.Log.Error("Failed to send outbox message to Kafka, ID: ", msg.ID, ", error: ", err)
			// Se l'invio fallisce, non facciamo nulla. Il messaggio rimane 'pending'
			// e verrà ritentato al prossimo ciclo.
			continue
		}

		// 4. Se l'invio ha successo, aggiorna lo stato nel database
		if err := storage.UpdateOutboxMessageStatus(ctx, msg.ID, "sent"); err != nil {
			logger.Log.Error("Failed to update outbox message status to 'sent', ID: ", msg.ID, ", error: ", err)
			// Questo è uno scenario critico: il messaggio è stato inviato ma non siamo riusciti
			// a marcare come tale. Questo causerà un reinvio, che dovrà essere gestito
			// dal consumatore (idempotenza).
			continue
		}

		logger.Log.Info("Successfully dispatched outbox message, ID: ", msg.ID)
	}
}
