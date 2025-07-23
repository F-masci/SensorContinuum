package main

import (
	"SensorContinuum/internal/edge-hub"
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"os"
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

	// Avvia la funzione per aggregare i dati.
	edge_hub.AggregateSensorData("6da1030b-2ba4-460d-967d-34d3b99dd4d7")

	logger.Log.Info("Edge Hub is terminating")
	logger.Log.Info("Shutting down Edge Hub...")
}
