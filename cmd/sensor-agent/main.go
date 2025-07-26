package main

import (
	"SensorContinuum/internal/sensor-agent/comunication"
	"SensorContinuum/internal/sensor-agent/environment"
	"SensorContinuum/internal/sensor-agent/health"
	"SensorContinuum/internal/sensor-agent/simulation"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"os"
)

// getContext ritorna il contesto del logger con le informazioni specifiche dell'agente del sensore
func getContext() logger.Context {
	return logger.Context{
		"service":  "sensor-agent",
		"building": environment.BuildingID,
		"floor":    environment.FloorID,
		"sensor":   environment.SensorID,
		"type":     environment.SensorType,
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
	logger.Log.Info("Sensor Location: ", environment.SensorLocation)
	logger.Log.Info("Sensor Type: ", environment.SensorType)
	logger.Log.Info("Sensor Reference: ", environment.SimulationSensorReference)

	// Inizializza la comunicazione con il simulatore del sensore
	sensorChannelSource := make(chan structure.SensorData, 100)
	go simulation.SimulateForever(sensorChannelSource)

	sensorChannelTarget := make(chan structure.SensorData, 100)
	// Invia i dati al broker MQTT
	go comunication.PublishData(sensorChannelTarget)

	go func() {
		for data := range sensorChannelSource {
			health.UpdateLastValueTimestamp()
			// Invia i dati al canale di comunicazione
			select {
			case sensorChannelTarget <- data:
				health.UpdateLastValueTimestamp()
			default:
				logger.Log.Warn("MQTT channel is full, discarding data: ", data)
			}
		}
	}()

	// Abilita il canale di comunicazione per health check
	if err := health.StartHealthCheckServer(":8080"); err != nil {
		logger.Log.Error("Failed to enable health check channel: ", err.Error())
		os.Exit(1)
	}

}
