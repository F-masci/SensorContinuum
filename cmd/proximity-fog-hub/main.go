package main

import (
	"SensorContinuum/internal/proximity-fog-hub"
	"SensorContinuum/internal/proximity-fog-hub/aggregation"
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
	"os"
	"time"
)

func main() {
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	logger.CreateLogger(logger.GetProximityHubContext(environment.EdgeMacrozone, environment.HubID))
	logger.PrintCurrentLevel()
	logger.Log.Info("Starting Proximity Fog Hub...")

	// Invia il messaggio di configurazione al Region Hub
	if err := comunication.SendRegistrationMessage(); err != nil {
		logger.Log.Error("Failed to send configuration message to Region Hub, error: ", err)
		os.Exit(1)
	}
	logger.Log.Info("Configuration message sent to Intermediate Fog Hub successfully.")

	// Connessione al DB per la cache
	if err := storage.InitDatabaseConnection(); err != nil {
		logger.Log.Error("failed to connect with local db, error: ", err)
		os.Exit(1)
	}

	filteredDataChannel := make(chan types.SensorData, 1000)
	configurationMessageChannel := make(chan types.ConfigurationMsg, 100)
	comunication.SetupMQTTConnection(filteredDataChannel, configurationMessageChannel)

	go proximity_fog_hub.ProcessEdgeHubData(filteredDataChannel)

	go proximity_fog_hub.ProcessEdgeHubConfiguration(configurationMessageChannel)

	// Avvio del ticker per l'aggregazione periodica, per ora metto ogni 2 minuti ma poi passa a ogni 5
	statsTicker := time.NewTicker(2 * time.Minute)
	logger.Log.Info("Ticker started, aggregating data every 2 minutes from now.")
	defer statsTicker.Stop()

	go func() {
		// Esegui subito la prima volta per non aspettare 5 minuti
		aggregation.PerformAggregationAndSend()
		// Iniziamo poi con il loop infinito
		for {
			select {
			case <-statsTicker.C:
				aggregation.PerformAggregationAndSend()
			}
		}
	}()

	logger.Log.Info("Proximity Fog Hub is running. Waiting for termination signal (Ctrl+C)...")
	utils.WaitForTerminationSignal()
	logger.Log.Info("Shutting down Edge Hub...")
}
