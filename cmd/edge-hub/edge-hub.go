package main

import (
	"SensorContinuum/internal/edge-hub"
	"SensorContinuum/internal/edge-hub/comunication"
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"SensorContinuum/pkg/utils"
	"os"
	"time"
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

	// Avvia il produttore (MQTT)
	sensorDataChannel := make(chan structure.SensorData, 100)
	comunication.SetupMQTTConnection(sensorDataChannel)

	// Avvia il filtro in un'altra goroutine.
	go edge_hub.FilterSensorData(sensorDataChannel)

	aggregateTicker := time.NewTicker(time.Minute)
	defer aggregateTicker.Stop()

	filteredDataChannel := make(chan structure.SensorData, 100)
	go comunication.PublishFilteredData(filteredDataChannel)

	go func() {
		for {
			select {
			case <-aggregateTicker.C:
				edge_hub.AggregateAllSensorsData(filteredDataChannel)
			}
		}
	}()

	cleanHealthTicker := time.NewTicker(time.Minute)
	defer cleanHealthTicker.Stop()

	go func() {
		for {
			select {
			case <-cleanHealthTicker.C:
				unhealthySensors := edge_hub.CleanUnhealthySensors()
				edge_hub.NotifyUnhealthySensors(unhealthySensors)
			}
		}
	}()

	utils.WaitForTerminationSignal()

	logger.Log.Info("Edge Hub is terminating")
	logger.Log.Info("Shutting down Edge Hub...")
}
