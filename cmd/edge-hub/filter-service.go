package main

import (
	"SensorContinuum/internal/edge-hub"
	"SensorContinuum/internal/edge-hub/comunication"
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"SensorContinuum/pkg/utils"
	"os"
)

func main() {
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}
	const service = "edge-hub-filter"
	logger.CreateLogger(logger.GetContext(service, environment.BuildingID, environment.FloorID, environment.HubID))
	logger.Log.Info("Starting Edge Hub - Filtering service...")

	// Avvia il produttore (MQTT)
	sensorDataChannel := make(chan structure.SensorData, 100)
	comunication.SetupMQTTConnection(sensorDataChannel)

	// Avvia il filtro in un'altra goroutine.
	go edge_hub.FilterSensorData(sensorDataChannel)

	logger.Log.Info("Edge Hub is running. Waiting for termination signal (Ctrl+C)...")

	// creazione canale quit che attende segnali per terminare in modo controllato
	utils.WaitForTerminationSignal()

	logger.Log.Info("Shutting down Edge Hub - Filtering service...")
}
