package cleaner

import (
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"context"
	"time"
)

const (
	// cleanerInterval definisce ogni quanto il cleaner si attiva per pulire la tabella outbox.
	// per ora lo metto ad ogni 5 minuti, poi in fase di deploy lo mandiamo anche a 6 ore o 1 giornp
	cleanerInterval = 5 * time.Minute
	// sentMessageMaxAge definisce l'età minima che un messaggio 'sent' deve avere prima di essere eliminato.
	// Questo fornisce una finestra di sicurezza per il debug o l'auditing, evitando di cancellare
	// messaggi che sono stati appena inviati. Per ora la metto a 1 ora
	sentMessageMaxAge = 1 * time.Hour
)

// Run avvia il processo di pulizia della tabella outbox, è molto simile a quella del dispatcher.
// Essa usa infatti un ticker e un ctx per uno spegnimento pulito
// Questa funzione viene eseguita in una goroutine separata e si occupa di:
//  1. Attivarsi periodicamente (ogni 5 minuti).
//  2. Cancellare i messaggi dalla tabella 'aggregated_stats_outbox' che sono in stato 'sent'
//     e sono più vecchi di una soglia definita (es. 1 ora).
func Run(ctx context.Context) {
	logger.Log.Info("Starting Outbox Cleaner...")
	ticker := time.NewTicker(cleanerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Stopping Outbox Cleaner...")
			return
		case <-ticker.C:
			logger.Log.Info("Outbox Cleaner running...")
			cleanupSentMessages(ctx)
		}
	}
}

// cleanupSentMessages si occupa della logica effettiva di cancellazione.
func cleanupSentMessages(ctx context.Context) {
	deletedCount, err := storage.DeleteSentOutboxMessages(ctx, sentMessageMaxAge)
	if err != nil {
		logger.Log.Error("Error cleaning up sent outbox messages: ", err)
		return
	}
	// Logghiamo quanti messaggi sono stati cancellati
	if deletedCount > 0 {
		logger.Log.Info("Outbox Cleaner successfully deleted ", deletedCount, " sent messages.")
	} else {
		logger.Log.Info("Outbox Cleaner found no old sent messages to delete.")
	}
}
