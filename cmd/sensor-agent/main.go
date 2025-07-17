package main

import (
	"SensorContinuum/internal/sensor-agent"
	"SensorContinuum/internal/sensor-agent/comunication"
	"SensorContinuum/internal/sensor-agent/simulation"
	"SensorContinuum/pkg/logger"
	"os"
)

// getContext ritorna il contesto del logger con le informazioni specifiche dell'agente del sensore
func getContext() logger.Context {
	return logger.Context{
		"service":  "sensor-agent",
		"building": sensor_agent.BuildingID,
		"floor":    sensor_agent.FloorID,
		"sensor":   sensor_agent.SensorID,
	}
}

func main() {

	// Inizializza l'ambiente
	if err := sensor_agent.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	// Inizializza il logger con il contesto
	logger.CreateLogger(getContext())
	logger.Log.Info("Starting Sensor Agent...")
	logger.Log.Info("Building ID: ", sensor_agent.BuildingID)
	logger.Log.Info("Floor ID: ", sensor_agent.FloorID)
	logger.Log.Info("Sensor ID: ", sensor_agent.SensorID)

	// Inizializza la comunicazione con il simulatore del sensore
	sensorChannel := make(chan float64, 100)
	go simulation.SimulateForever(sensorChannel)

	// Invia i dati al broker MQTT
	for data := range sensorChannel {
		comunication.PublishData(data)
	}

}
