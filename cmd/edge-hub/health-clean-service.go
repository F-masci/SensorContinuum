package main

import (
	"SensorContinuum/internal/edge-hub"
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/utils"
	"os"
	"time"
)

func getContext() logger.Context {
	return logger.Context{
		"service":  "edge-hub-health-clean",
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
	logger.Log.Info("Starting Edge Hub - Health&Clean service...")

	if environment.OperationMode == "loop" {
		logger.Log.Info("Operation mode is set to 'loop'. The service will run indefinitely.")

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		go func() {
			for {
				select {
				case <-ticker.C:
					unhealthySensors := edge_hub.CleanUnhealthySensors()
					edge_hub.NotifyUnhealthySensors(unhealthySensors)
				}
			}
		}()

	}

	if environment.OperationMode == "once" {
		logger.Log.Info("Operation mode is set to 'once'. The service will run once and then terminate.")
		unhealthySensors := edge_hub.CleanUnhealthySensors()
		edge_hub.NotifyUnhealthySensors(unhealthySensors)
	}

	// creazione canale quit che attende segnali per terminare in modo controllato
	utils.WaitForTerminationSignal()

	logger.Log.Info("Shutting down Edge Hub - Health&Clean service...")
}
