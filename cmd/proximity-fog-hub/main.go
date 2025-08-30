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

/*	---------------------------- DESCRIZIONE  PROXIMITY FOG HUB ---------------------------------------------------------

il proximity fog hub svolge i seguenti compiti:

	- riceve i dati filtrati dall edge-hub tramite broker mqtt iscrivendosi allo stesso topic
	- salva i dati ricevuti in una cache locale (che mantiene i dati delle ultime 6 ore)
	- invia i dati ricevuti tramite kafka all' intermediate-fog-hub, in questo modo l' intermediate-fog-hub
      ha una copia dettagliata e immediata di ogni singolo dato e può rispondere alla domanda "cosa accade ora nel sistema?"
	- ogni 5 minuti (da modificare nel caso) scatta un ticker che:
		1) esegue una query sulla cache locale temporanea (sul db locale quindi) chiedendo di restituire i valori
            di max, min e avg di tutti i dati ricevuti negli ultimi 5 minuti
		2) riceve i risultati dal db
		3) salva queste statistiche in una tabella outbox locale, ci penserà poi un componente chiamato
			dispatcher (una goroutine che si avvia anche essa grazie a un ticker)
			a inviarle all' intermediate-fog-hub e settare lo stato 'sent' alla tabella outbox.
		4) un cleaner che si avvia sempre con un ticker in una goroutine separata si occupa di pulire la tabella outbox
			ossia nello specifico di eliminare i messaggi che sono in stato 'sent' e che hanno più di 1 ora (per il momento)
			per evitare che la tabella outbox cresca all'infinito

quindi alla fine saremo sicuri che l 'intermediate-fog-hub possieda una visione riassuntiva e
di più alto livello dello stato dell'edificio

------------------------------------------------------------------------------------------------------------------------- */

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
