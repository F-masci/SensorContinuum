package main

import (
	"SensorContinuum/internal/edge-hub"
	"SensorContinuum/internal/edge-hub/comunication"
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"os"
	"os/signal"
	"syscall"
)

func getContext() logger.Context {
	return logger.Context{
		"service":  "edge-hub",
		"building": environment.BuildingID,
		"floor":    environment.FloorID,
		"hub":      environment.HubID,
	}
}

func main() {
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	logger.CreateLogger(getContext())
	logger.Log.Info("Starting Edge Hub...")

	dataChannel := make(chan structure.SensorData, 100)

	// Avvia il produttore (MQTT) in una goroutine.
	// questo nuovo approccio prevede che la funzione sia bloccante
	// e che non ritornerà mai se non per un errore fatale.
	go comunication.PullSensorData(dataChannel)

	// Avvia il consumatore in un'altra goroutine.
	go edge_hub.ProcessSensorData(dataChannel)

	logger.Log.Info("Edge Hub is running. Waiting for termination signal (Ctrl+C)...")

	// creazione canale quit che attende segnali per terminare in modo controllato
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down Edge Hub...")
}
