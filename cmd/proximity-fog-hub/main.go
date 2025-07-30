package main

import (
	proximity_fog_hub "SensorContinuum/internal/proximity-fog-hub"
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"SensorContinuum/pkg/utils"
	"os"
)

// getContext ritorna il contesto del logger con le informazioni specifiche dell'agente del sensore
func getContext() logger.Context {
	return logger.Context{
		"service":  "proximity-fog-hub",
		"building": environment.BuildingID,
		"hub":      environment.HubID,
	}
}

func main() {

	// Inizializza l'ambiente
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	// Inizializza il logger con il contesto
	logger.CreateLogger(getContext())
	logger.Log.Info("Starting Proximity Fog Hub...")
	logger.Log.Info("Building ID: ", environment.BuildingID)
	logger.Log.Info("Hub ID: ", environment.HubID)

	filteredDataChannel := make(chan structure.SensorData, 100)
	// inizializza connessione MQTT in maniera sincrona
	comunication.SetupMQTTConnection(filteredDataChannel)

	go proximity_fog_hub.ProcessEdgeHubData(filteredDataChannel)

	logger.Log.Info("Proximity Fog Hub is running. Waiting for termination signal (Ctrl+C)...")
	utils.WaitForTerminationSignal()
	logger.Log.Info("Shutting down Edge Hub...")
}
