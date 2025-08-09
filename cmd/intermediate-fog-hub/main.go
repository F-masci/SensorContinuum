package main

import (
	"SensorContinuum/internal/intermediate-fog-hub"
	"SensorContinuum/internal/intermediate-fog-hub/comunication"
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
	"os"
)

// getContext ritorna il contesto del logger con le informazioni specifiche dell'agente del sensore
func getContext() logger.Context {
	return logger.Context{
		"service":   "intermediate-fog-hub",
		"macrozone": environment.EdgeMacrozone,
		"hub":       environment.HubID,
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
	intermediate_fog_hub.Register()
	logger.Log.Info("Building ID: ", environment.EdgeMacrozone)
	logger.Log.Info("Hub ID: ", environment.HubID)

	dataChannel := make(chan types.SensorData)
	go func() {
		// Se la funzione ritorna (a causa di un errore), lo logghiamo.
		// Questo farà terminare l'applicazione.
		err := comunication.PullAggregatedData(dataChannel)
		if err != nil {
			logger.Log.Error("Kafka consumer has stopped", "error", err.Error())
			os.Exit(1)
		}
	}()

	msgChannel := make(chan types.ConfigurationMsg)
	go func() {
		// Se la funzione ritorna (a causa di un errore), lo logghiamo.
		// Questo farà terminare l'applicazione.
		err := comunication.PullConfigurationMessage(msgChannel)
		if err != nil {
			logger.Log.Error("Kafka consumer has stopped", "error", err.Error())
			os.Exit(1)
		}
	}()

	// Avvia il processo di gestione dei dati intermedi
	go intermediate_fog_hub.ProcessProximityFogHubData(dataChannel)

	// Avvia il processo di gestione dei messaggi di configurazione
	go intermediate_fog_hub.ProcessProximityFogHubConfiguration(msgChannel)

	logger.Log.Info("Intermediate Fog Hub is running. Waiting for termination signal (Ctrl+C)...")
	utils.WaitForTerminationSignal()
	logger.Log.Info("Shutting down Intermediate Fog Hub...")

}
