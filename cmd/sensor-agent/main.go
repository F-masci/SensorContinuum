package main

import (
	"SensorContinuum/internal/sensor-agent/comunication"
	"SensorContinuum/internal/sensor-agent/environment"
	"SensorContinuum/internal/sensor-agent/simulation"
	"SensorContinuum/pkg/logger"
	"os"
)

// getContext ritorna il contesto del logger con le informazioni specifiche dell'agente del sensore
func getContext() logger.Context {
	return logger.Context{
		"service":  "sensor-agent",
		"building": environment.BuildingID,
		"floor":    environment.FloorID,
		"sensor":   environment.SensorID,
	}
}

func main() {

	// Inizializza l'ambiente
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	// Inizializza il logger con il contesto
	logger.CreateLogger(getContext())
	logger.Log.Info("Starting Sensor Agent...")
	logger.Log.Info("Building ID: ", environment.BuildingID)
	logger.Log.Info("Floor ID: ", environment.FloorID)
	logger.Log.Info("Sensor ID: ", environment.SensorID)

	// Inizializza la comunicazione con il simulatore del sensore
	sensorChannel := make(chan float64, 100)
	go simulation.SimulateForever(sensorChannel)

	// Invia i dati al broker MQTT
	for data := range sensorChannel {
		comunication.PublishData(data)
	}

}
