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

	dataChannel := make(chan structure.SensorData, 100)

	// Avvia il produttore (MQTT) in una goroutine.
	// questo nuovo approccio prevede che la funzione sia bloccante
	// e che non ritornerà mai se non per un errore fatale.
	go comunication.PullSensorData(dataChannel)

	// Avvia il filtro in un'altra goroutine.
	go edge_hub.FilterSensorData(dataChannel)

	aggregateTicker := time.NewTicker(time.Minute)
	defer aggregateTicker.Stop()

	go func() {
		for {
			select {
			case <-aggregateTicker.C:
				edge_hub.AggregateAllSensorsData()
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
