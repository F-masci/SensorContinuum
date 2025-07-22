package main

import (
	proximity_fog_hub "SensorContinuum/internal/proximity-fog-hub"
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"os"
	"os/signal"
	"syscall"
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

	dataChannel := make(chan structure.SensorData)
	go comunication.PullFilteredData(dataChannel)

	go proximity_fog_hub.ProcessEdgeHubData(dataChannel)

	logger.Log.Info("Proximity Fog Hub is running. Waiting for termination signal (Ctrl+C)...")

	// creazione canale quit che attende segnali per terminare in modo controllato
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down Edge Hub...")
}
