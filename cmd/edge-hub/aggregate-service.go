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

func main() {
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}
	const service = "edge-hub-aggregate"
	logger.CreateLogger(logger.GetContext(service, environment.BuildingID, environment.FloorID, environment.HubID))
	logger.Log.Info("Starting Edge Hub - Aggregate service...")

	filteredDataChannel := make(chan structure.SensorData, 100)
	go comunication.PublishFilteredData(filteredDataChannel)

	if environment.OperationMode == "loop" {
		logger.Log.Info("Operation mode is set to 'loop'. The service will run indefinitely.")

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		go func() {
			for {
				select {
				case <-ticker.C:
					edge_hub.AggregateAllSensorsData(filteredDataChannel)
				}
			}
		}()

	}

	if environment.OperationMode == "once" {
		logger.Log.Info("Operation mode is set to 'once'. The service will run once and then terminate.")
		edge_hub.AggregateAllSensorsData(filteredDataChannel)
		logger.Log.Info("Aggregation completed. The service will now terminate.")
		os.Exit(0)
	}

	utils.WaitForTerminationSignal()

	logger.Log.Info("Edge Hub is terminating")
	logger.Log.Info("Shutting down Edge Hub - Aggregate service...")
}
