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
	Per ora l' intermediate fog hub:

- riceve i dati che l edge-hub invia al proximity-fog-hub in modo da poter rispondere alla domanda "cosa accade ora?"
- riceve i dati aggregati ( le statistiche ) ogni tot minuti dal proximity fog hub tramite kafka
- il suo compito è quindi quello di storage di dati sia dettagliati ( i dati che arrivano in tempo reale ) sia aggregati ( le statistiche che arrivano ogni tot minuti )
- altre responsabilità le implementerò in futuro
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
		go intermediate_fog_hub.ProcessRealTimeData(realTimeDataChannel)

		go func() {
			// Se la funzione ritorna (a causa di un errore), lo logghiamo.
			// Questo farà terminare l'applicazione.
			err := comunication.PullRealTimeData(realTimeDataChannel)
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
		go intermediate_fog_hub.ProcessStatisticsData(statsDataChannel)

		go func() {
			err := comunication.PullStatisticsData(statsDataChannel)
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
		configurationMessageChannel := make(chan types.ConfigurationMsg)
		go intermediate_fog_hub.ProcessProximityFogHubConfiguration(configurationMessageChannel)

		go func() {
			// Se la funzione ritorna (a causa di un errore), lo logghiamo.
			// Questo farà terminare l'applicazione.
			err := comunication.PullConfigurationMessage(configurationMessageChannel)
			if err != nil {
				logger.Log.Error("Kafka consumer for configuration message has stopped: ", err.Error())
				os.Exit(1)
			}
		}()

	}

	/* -------- HEARTBEAT SERVICE -------- */

	if environment.ServiceMode == types.IntermediateHubHeartbeatService || environment.ServiceMode == types.IntermediateHubService {

		// Avvia il processo di gestione dei messaggi di heartbeat
		heartbeatChannel := make(chan types.HeartbeatMsg)
		go intermediate_fog_hub.ProcessProximityFogHubHeartbeat(heartbeatChannel)

		go func() {
			// Se la funzione ritorna (a causa di un errore), lo logghiamo.
			// Questo farà terminare l'applicazione.
			err := comunication.PullHeartbeatMessage(heartbeatChannel)
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
