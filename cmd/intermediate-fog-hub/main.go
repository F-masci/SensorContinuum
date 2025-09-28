package main

import (
	"SensorContinuum/configs/timeouts"
	"SensorContinuum/internal/intermediate-fog-hub"
	"SensorContinuum/internal/intermediate-fog-hub/aggregation"
	"SensorContinuum/internal/intermediate-fog-hub/comunication"
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/internal/intermediate-fog-hub/health"
	"SensorContinuum/internal/intermediate-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
	"context"
	"os"
	"time"
)

/*
DESCRIZIONE FUNZIONALE:
L'Intermediate Hub è il punto di convergenza regionale, focalizzato sull'ingestione ad alto volume, l'aggregazione finale e la persistenza a lungo termine dei dati elaborati.

RESPONSABILITÀ CHIAVE:

1.  Ingestione e Ottimizzazione I/O: I Servizi di Data Ingestion leggono i messaggi da Kafka e li raccolgono in batch in memoria. Questi batch sono scritti in blocco nel database a lungo termine (PostgreSQL) utilizzando l'operazione COPY FROM per ottimizzare l'efficienza.

2.  Gestione Affidabile dell'Offset: Viene implementata la gestione manuale dell'offset di Kafka. L'offset viene contrassegnato come letto solo dopo il successo della scrittura nel database, garantendo l'integrità dei dati e la semantica at-least-once.

3.  Controllo del Flusso: La lettura da Kafka viene temporaneamente sospesa durante le operazioni di scrittura sul database per prevenire il sovraccarico e l'accumulo di backlog.

4.  Aggregazione Finale: Calcola le statistiche aggregate a livello di regione, sfruttando i dati pre-elaborati (statistiche di macrozona) ricevuti dai livelli inferiori.

5.  Gestione Metadati: Aggiorna i timestamp dell'ultima comunicazione dei sensori contestualmente all'inserimento dei batch di dati nel database.
*/
func main() {

	// Setup dell'ambiente
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	// Inizializza il logger
	logger.CreateLogger(logger.GetIntermediateHubContext(environment.HubID))
	logger.PrintCurrentLevel()
	logger.Log.Info("Starting Intermediate Fog Hub...")

	// Si registra nel sistema
	logger.Log.Info("Registering the intermediate fog hub...")
	if err := storage.SelfRegistration(); err != nil {
		logger.Log.Error("Failed to register the intermediate fog hub: ", err)
		os.Exit(1)
	}
	logger.Log.Info("Intermediate fog hub registered successfully.")

	// Avvia il thread per l'aggiornamento del last seen
	go func() {
		for {
			err := storage.UpdateLastSeenRegionHub()
			if err != nil {
				logger.Log.Error("Failed to update last seen timestamp: ", err)
				os.Exit(1)
			}
			logger.Log.Info("Updated last seen timestamp")
			time.Sleep(timeouts.HeartbeatInterval)
		}
	}()

	/* -------- REAL-TIME SERVICE -------- */

	if environment.ServiceMode == types.IntermediateHubRealtimeService || environment.ServiceMode == types.IntermediateHubService {

		// Avvia il processo di gestione dei dati intermedi
		realTimeDataChannel := make(chan types.SensorData, environment.SensorDataBatchSize*3)
		realTimePauseSignal := utils.NewPauseSignal()
		go intermediate_fog_hub.ProcessRealTimeData(realTimeDataChannel, realTimePauseSignal)

		go func() {
			// Se la funzione ritorna (a causa di un errore), lo logghiamo.
			// Questo farà terminare l'applicazione.
			err := comunication.PullRealTimeData(realTimeDataChannel, realTimePauseSignal)
			if err != nil {
				logger.Log.Error("Kafka consumer for the real time data has stopped: ", err.Error())
				os.Exit(1)
			}
		}()

	}

	/* -------- STATISTICS SERVICE -------- */

	if environment.ServiceMode == types.IntermediateHubStatisticsService || environment.ServiceMode == types.IntermediateHubService {

		// Avvia il processo di gestione dei dati statistici
		statsDataChannel := make(chan types.AggregatedStats, environment.AggregatedDataBatchSize*3)
		zoneStatsPauseSignal := utils.NewPauseSignal()
		macrozoneStatsPauseSignal := utils.NewPauseSignal()
		go intermediate_fog_hub.ProcessStatisticsData(statsDataChannel, zoneStatsPauseSignal, macrozoneStatsPauseSignal)

		go func() {
			err := comunication.PullStatisticsData(statsDataChannel, zoneStatsPauseSignal, macrozoneStatsPauseSignal)
			if err != nil {
				// Se la funzione ritorna (a causa di un errore), lo logghiamo.
				// Questo farà terminare l'applicazione.
				logger.Log.Error("Kafka consumer for statistics has stopped: ", err)
				os.Exit(1)
			}
		}()

	}

	/* -------- CONFIGURATION SERVICE -------- */

	if environment.ServiceMode == types.IntermediateHubConfigurationService || environment.ServiceMode == types.IntermediateHubService {

		// Avvia il processo di gestione dei messaggi di configurazione
		configurationMessageChannel := make(chan types.ConfigurationMsg, environment.ConfigurationMessageBatchSize*3)
		configurationPauseSignal := utils.NewPauseSignal()
		go intermediate_fog_hub.ProcessProximityFogHubConfiguration(configurationMessageChannel, configurationPauseSignal)

		go func() {
			// Se la funzione ritorna (a causa di un errore), lo logghiamo.
			// Questo farà terminare l'applicazione.
			err := comunication.PullConfigurationMessage(configurationMessageChannel, configurationPauseSignal)
			if err != nil {
				logger.Log.Error("Kafka consumer for configuration message has stopped: ", err.Error())
				os.Exit(1)
			}
		}()

	}

	/* -------- HEARTBEAT SERVICE -------- */

	if environment.ServiceMode == types.IntermediateHubHeartbeatService || environment.ServiceMode == types.IntermediateHubService {

		// Avvia il processo di gestione dei messaggi di heartbeat
		heartbeatChannel := make(chan types.HeartbeatMsg, environment.HeartbeatMessageBatchSize*3)
		heartbeatPauseSignal := utils.NewPauseSignal()
		go intermediate_fog_hub.ProcessProximityFogHubHeartbeat(heartbeatChannel, heartbeatPauseSignal)

		go func() {
			// Se la funzione ritorna (a causa di un errore), lo logghiamo.
			// Questo farà terminare l'applicazione.
			err := comunication.PullHeartbeatMessage(heartbeatChannel, heartbeatPauseSignal)
			if err != nil {
				logger.Log.Error("Kafka consumer for heartbeat has stopped: ", err.Error())
				os.Exit(1)
			}
		}()

	}

	/* -------- AGGREGATOR SERVICE -------- */

	if (environment.ServiceMode == types.IntermediateHubAggregatorService && environment.OperationMode == types.OperationModeLoop) || environment.ServiceMode == types.IntermediateHubService {
		// Avvia il servizio di aggregazione in una goroutine separata.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		aggregation.Run(ctx)
	}

	if environment.ServiceMode == types.IntermediateHubAggregatorService && environment.OperationMode == types.OperationModeOnce {
		// Esegue una singola aggregazione e termina.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		aggregation.AggregateSensorData(ctx)
		logger.Log.Info("Aggregation completed. The service will now terminate.")
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

	logger.Log.Info("Intermediate Fog Hub is running. Waiting for termination signal (Ctrl+C)...")
	utils.WaitForTerminationSignal()
	logger.Log.Info("Shutting down Intermediate Fog Hub...")

}
