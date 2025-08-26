package main

import (
	"SensorContinuum/internal/intermediate-fog-hub"
	"SensorContinuum/internal/intermediate-fog-hub/aggregation"
	"SensorContinuum/internal/intermediate-fog-hub/comunication"
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/internal/intermediate-fog-hub/health"
	"SensorContinuum/internal/intermediate-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
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

	// Inizializza l'ambiente
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		return
	}

	// Inizializza il logger con il contesto
	logger.CreateLogger(logger.GetIntermediateHubContext(environment.HubID))
	logger.PrintCurrentLevel()
	logger.Log.Info("Starting Intermediate Fog Hub...")

	storage.Register()

	// Avvia il processo di gestione dei dati intermedi
	realTimeDataChannel := make(chan types.SensorData)
	go intermediate_fog_hub.ProcessRealTimeData(realTimeDataChannel)

	// Avvia il processo di gestione dei dati statistici
	statsDataChannel := make(chan types.AggregatedStats)
	go intermediate_fog_hub.ProcessStatisticsData(statsDataChannel)

	// Avvia il processo di gestione dei messaggi di configurazione
	configurationMessageChannel := make(chan types.ConfigurationMsg)
	go intermediate_fog_hub.ProcessProximityFogHubConfiguration(configurationMessageChannel)

	// Avvia il processo di gestione dei messaggi di heartbeat
	heartbeatChannel := make(chan types.HeartbeatMsg)
	go intermediate_fog_hub.ProcessProximityFogHubHeartbeat(heartbeatChannel)

	go func() {
		// Se la funzione ritorna (a causa di un errore), lo logghiamo.
		// Questo farà terminare l'applicazione.
		err := comunication.PullRealTimeData(realTimeDataChannel)
		if err != nil {
			logger.Log.Error("Kafka consumer for the real time data has stopped: ", err.Error())
			os.Exit(1)
		}
	}()

	go func() {
		err := comunication.PullStatisticsData(statsDataChannel)
		if err != nil {
			// Se la funzione ritorna (a causa di un errore), lo logghiamo.
			// Questo farà terminare l'applicazione.
			logger.Log.Error("Kafka consumer for statistics has stopped: ", err)
			os.Exit(1)
		}
	}()

	go func() {
		// Se la funzione ritorna (a causa di un errore), lo logghiamo.
		// Questo farà terminare l'applicazione.
		err := comunication.PullConfigurationMessage(configurationMessageChannel)
		if err != nil {
			logger.Log.Error("Kafka consumer for configuration message has stopped: ", err.Error())
			os.Exit(1)
		}
	}()

	go func() {
		// Se la funzione ritorna (a causa di un errore), lo logghiamo.
		// Questo farà terminare l'applicazione.
		err := comunication.PullHeartbeatMessage(heartbeatChannel)
		if err != nil {
			logger.Log.Error("Kafka consumer for heartbeat has stopped: ", err.Error())
			os.Exit(1)
		}
	}()

	// Avvio del ticker per l'aggregazione periodica, per ora metto ogni 2 minuti ma poi passa a ogni 5
	statsTicker := time.NewTicker(2 * time.Minute)
	logger.Log.Info("Ticker started, aggregating data every 2 minutes from now.")
	defer statsTicker.Stop()

	go func() {

		for {
			select {
			case <-statsTicker.C:
				aggregation.PerformAggregationAndSave()
			}
		}
	}()

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
