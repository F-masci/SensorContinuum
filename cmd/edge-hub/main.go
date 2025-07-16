package main

import (
	"SensorContinuum/internal/edge-hub"
	"SensorContinuum/internal/edge-hub/comunication"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"os"
)

// getContext ritorna il contesto del logger con le informazioni specifiche dell'hub
func getContext() logger.Context {
	return logger.Context{
		"service":  "edge-hub",
		"building": edge_hub.BuildingID,
		"floor":    edge_hub.FloorID,
		"hub":      edge_hub.HubID,
	}
}

func pullSensorData(dataChannel chan structure.SensorData) {
	err := comunication.PullSensorData(dataChannel)
	if err != nil {
		logger.Log.Error("Failed to pull sensor data:", err.Error())
		return
	}
}

func main() {

	// Inizializza l'ambiente
	if err := edge_hub.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	// Inizializza il logger con il contesto
	logger.CreateLogger(getContext())
	logger.Log.Info("Starting Edge Hub...")

	// Avvia il servizio di processamento dei dati
	logger.Log.Info("Starting data processing...")

	for {
		logger.Log.Debug("Waiting for data...")

		// Ottieni i dati dai sensori
		dataChannel := make(chan structure.SensorData)
		go pullSensorData(dataChannel)

		// Processa i dati dei sensori
		edge_hub.ProcessSensorData(dataChannel)
	}

}
