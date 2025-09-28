package main

import (
	proximity_fog_hub "SensorContinuum/internal/proximity-fog-hub"
	"SensorContinuum/internal/proximity-fog-hub/aggregation"
	"SensorContinuum/internal/proximity-fog-hub/cleaner"
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/dispatcher"
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/internal/proximity-fog-hub/health"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
	"context"
	"os"
)

/*
DESCRIZIONE FUNZIONALE:
Il Proximity Hub è il nodo di aggregazione Fog critico, responsabile della persistenza temporanea dei dati e della garanzia di inoltro affidabile verso il livello intermedio (Cloud). Utilizza un database PostgreSQL locale.

RESPONSABILITÀ CHIAVE:

1.  Persistenza Idempotente: Il Modulo Local Cache salva i dati aggregati ricevuti via MQTT in un database PostgreSQL locale con stato "pending". Implementa una logica di de-duplicazione a livello di database per assicurare l'idempotenza.

2.  Aggregazione Statistica: Il Modulo Aggregator calcola le statistiche (max, min, avg) a livello di zona e macrozona. Elabora un intervallo di 15 minuti con un offset di 10 minuti per la gestione della latenza e ha la capacità di colmare le lacune nelle aggregazioni.

3.  Affidabilità (Transactional Outbox): Il Modulo Dispatcher implementa il pattern Transactional Outbox. Preleva i messaggi "pending" dal database e li inoltra in modo affidabile al broker Kafka (Intermediate Hub), aggiornando lo stato a "sent" solo dopo il successo dell'invio.

4.  Proxy di Controllo: I Moduli Proxy Heartbeat e Proxy Configuration raccolgono i messaggi di controllo dagli Edge Hub (via MQTT) e li inoltrano su Kafka.

5.  Manutenzione: Il Modulo Cleaner elimina periodicamente i messaggi con stato "sent" che superano il periodo di conservazione stabilito (tipicamente 1-2 giorni) per prevenire la crescita indefinita della tabella outbox.
*/
func main() {

	// Setup dell'ambiente
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	// Inizializza il logger
	logger.CreateLogger(logger.GetProximityHubContext(environment.EdgeMacrozone, environment.HubID))
	logger.PrintCurrentLevel()
	logger.Log.Info("Starting Proximity Fog Hub...")

	// Connessione al DB per la cache
	if err := storage.InitDatabaseConnection(); err != nil {
		logger.Log.Error("failed to connect with local db, error: ", err)
		os.Exit(1)
	}

	// Creazione dei canali per i messaggi di configurazione, heartbeat e dati filtrati
	filteredDataChannel := make(chan types.SensorData, 100)
	configurationMessageChannel := make(chan types.ConfigurationMsg, 100)
	heartbeatMessageChannel := make(chan types.HeartbeatMsg, 100)
	// Inizializza connessione MQTT in maniera sincrona
	comunication.SetupMQTTConnection(filteredDataChannel, configurationMessageChannel, heartbeatMessageChannel)

	// Si registra al proximity Hub in base al proprio Service Mode
	// Questo invio è sincrono, se fallisce l'applicazione termina
	logger.Log.Info("Sending own registration message")
	comunication.SendOwnRegistrationMessage()
	logger.Log.Info("Registration message sent successfully")

	// Avvia il thread per l'invio dei messaggi di heartbeat
	go comunication.SendOwnHeartbeatMessage()

	/* ----- LOCAL CACHE SERVICE ------ */

	if environment.ServiceMode == types.ProximityHubLocalCacheService || environment.ServiceMode == types.ProximityHubService {
		// Avvia l'elaborazione dei dati filtrati in un'altra goroutine.
		// Riceve i dati dal canale filteredDataChannel e li salva nella cache locale.
		go proximity_fog_hub.ProcessEdgeHubData(filteredDataChannel)
	}

	/* ----- CONFIGURATION SERVICE ------ */

	if environment.ServiceMode == types.ProximityHubConfigurationService || environment.ServiceMode == types.ProximityHubService {
		// Avvia l'elaborazione dei messaggi di configurazione in un'altra goroutine.
		// Riceve i messaggi dal canale configurationMessageChannel e li elabora.
		go proximity_fog_hub.ProcessEdgeHubConfiguration(configurationMessageChannel)
	}

	/* ----- HEARTBEAT SERVICE ------ */

	if environment.ServiceMode == types.ProximityHubHeartbeatService || environment.ServiceMode == types.ProximityHubService {
		// Avvia l'elaborazione dei messaggi di heartbeat in un'altra goroutine.
		// Riceve i messaggi dal canale heartbeatMessageChannel e li elabora.
		go proximity_fog_hub.ProcessEdgeHubHeartbeat(heartbeatMessageChannel)
	}

	/* ----- AGGREGATOR SERVICE ------ */

	if (environment.ServiceMode == types.ProximityHubAggregatorService && environment.OperationMode == types.OperationModeLoop) || environment.ServiceMode == types.ProximityHubService {
		// Avvia il servizio di aggregazione in una goroutine separata.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		aggregation.Run(ctx)
	}

	if environment.ServiceMode == types.ProximityHubAggregatorService && environment.OperationMode == types.OperationModeOnce {
		// Esegue una singola aggregazione e termina.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		aggregation.AggregateSensorData(ctx)
		logger.Log.Info("Aggregation completed. The service will now terminate.")
		os.Exit(0)
	}

	/* ---- DISPATCHER SERVICE ------ */

	if (environment.ServiceMode == types.ProximityHubDispatcherService && environment.OperationMode == types.OperationModeLoop) || environment.ServiceMode == types.ProximityHubService {
		// Avviamo il dispatcher in una goroutine separata.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		dispatcher.Run(ctx)
	}

	if environment.ServiceMode == types.ProximityHubDispatcherService && environment.OperationMode == types.OperationModeOnce {
		// Esegue una singola esecuzione del dispatcher e termina.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		dispatcher.ProcessPendingMessages(ctx)
		logger.Log.Info("Dispatcher completed. The service will now terminate.")
		os.Exit(0)
	}

	/* ---- CLEANER SERVICE ------ */

	if (environment.ServiceMode == types.ProximityHubCleanerService && environment.OperationMode == types.OperationModeLoop) || environment.ServiceMode == types.ProximityHubService {
		// Avviamo il cleaner in una goroutine separata.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go cleaner.Run(ctx)
	}

	if environment.ServiceMode == types.ProximityHubCleanerService && environment.OperationMode == types.OperationModeOnce {
		// Esegue una singola esecuzione del cleaner e termina.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cleaner.CleanupSentMessages(ctx)
		logger.Log.Info("Cleaner completed. The service will now terminate.")
		os.Exit(0)
	}

	/* -------- HEALTH CHECK SERVER -------- */

	if environment.HealthzServer {
		logger.Log.Info("Enabling health check channel on port " + environment.HealthzServerPort)
		go func() {
			if err := health.StartHealthCheckServer(":" + environment.HealthzServerPort); err != nil {
				logger.Log.Error("Failed to enable health check channel: ", err.Error())
				os.Exit(1)
			}
		}()
	}

	// Attende il segnale di terminazione (ad esempio Ctrl+C)
	utils.WaitForTerminationSignal()
	logger.Log.Info("Shutting down Proximity Fog Hub...")
}
