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

/*
	Per ora l' intermediate fog hub:

- riceve i dati che l edge-hub invia al proximity-fog-hub in modo da poter rispondere alla domanda "cosa accade ora?"
- riceve i dati aggregati ( le statistiche ) ogni tot minuti dal proximity fog hub tramite kafka
- il suo compito è quindi quello di storage di dati sia dettagliati ( i dati che arrivano in tempo reale ) sia aggregati ( le statistiche che arrivano ogni tot minuti )
- altre responsabilità le implementerò in futuro
*/
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

	realTimeDataChannel := make(chan structure.SensorData)
	go func() {
		// Se la funzione ritorna (a causa di un errore), lo logghiamo.
		// Questo farà terminare l'applicazione.
		err := comunication.PullRealTimeData(realTimeDataChannel)
		if err != nil {
			logger.Log.Error("Kafka consumer for the real time data has stopped", "error", err.Error())
			os.Exit(1)
		}
	}()

	// Avvia il processo di gestione dei dati intermedi
	go intermediate_fog_hub.ProcessRealTimeData(realTimeDataChannel)

	// Canale per i dati statistici
	statsDataChannel := make(chan structure.AggregatedStats)
	go func() {
		err := comunication.PullStatisticsData(statsDataChannel)
		if err != nil {
			logger.Log.Error("Consumatore Kafka per statistiche si è fermato", "error", err)
			os.Exit(1)
		}
	}()
	go intermediate_fog_hub.ProcessStatisticsData(statsDataChannel)

	logger.Log.Info("Intermediate Fog Hub is running. Waiting for termination signal (Ctrl+C)...")
	utils.WaitForTerminationSignal()
	logger.Log.Info("Shutting down Intermediate Fog Hub...")

}
