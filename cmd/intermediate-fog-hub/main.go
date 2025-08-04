package main

import (
	"SensorContinuum/internal/intermediate-fog-hub"
	"SensorContinuum/internal/intermediate-fog-hub/comunication"
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"SensorContinuum/pkg/utils"
	"os"
)

// getContext ritorna il contesto del logger con le informazioni specifiche dell'agente del sensore
func getContext() logger.Context {
	return logger.Context{
		"service":  "intermediate-fog-hub",
		"building": environment.BuildingID,
		"hub":      environment.HubID,
	}
}

func main() {

	// Inizializza l'ambiente
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		return
	}

	// Inizializza il logger con il contesto
	logger.CreateLogger(getContext())
	logger.Log.Info("Starting Intermediate Fog Hub...")
	logger.Log.Info("Building ID: ", environment.BuildingID)
	logger.Log.Info("Hub ID: ", environment.HubID)

	dataChannel := make(chan structure.SensorData)
	go func() {
		// Se la funzione ritorna (a causa di un errore), lo logghiamo.
		// Questo farà terminare l'applicazione.
		err := comunication.PullAggregatedData(dataChannel)
		if err != nil {
			logger.Log.Error("Kafka consumer has stopped", "error", err.Error())
			os.Exit(1)
		}
	}()

	// Avvia il processo di gestione dei dati intermedi
	go intermediate_fog_hub.ProcessProximityFogHubData(dataChannel)

	logger.Log.Info("Intermediate Fog Hub is running. Waiting for termination signal (Ctrl+C)...")
	utils.WaitForTerminationSignal()
	logger.Log.Info("Shutting down Intermediate Fog Hub...")

}
