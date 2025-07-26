package simulation

import (
	"SensorContinuum/configs/simulation"
	"SensorContinuum/pkg/structure"
	"os"
	"time"

	"SensorContinuum/pkg/logger"
)

const infiniteValue = -1234

func setupSimulator() error {

	// Scarica un file CSV random della data attuale meno 2 giorni
	filePath, err := downloadRandomCSV()
	if err != nil {
		logger.Log.Error("Error downloading random CSV: ", err)
		return err
	} else {
		logger.Log.Info("CSV file downloaded successfully: ", filePath)
	}

	// Inizializza la distribuzione con il file CSV scaricato
	err = setupDistribution(filePath)
	if err != nil {
		logger.Log.Error("Error setting up distribution: ", err)
		return err
	} else {
		logger.Log.Info("Distribution setup successfully with file: ", filePath)
	}

	return nil
}

func Simulate(nValue int, dataChannel chan structure.SensorData) error {

	err := setupSimulator()
	if err != nil {
		logger.Log.Error("Error setting up simulator: ", err)
		return err
	}

	if nValue == infiniteValue {
		logger.Log.Info("Simulating sensor data indefinitely...")
	} else {
		logger.Log.Info("Simulating sensor data for ", nValue, " values...")
	}

	for nValue == infiniteValue || nValue > 0 {

		sensorData := generateSensorData()

		logger.Log.Debug("Sensor reading: ", sensorData.Data)

		if dataChannel != nil && sensorData != (structure.SensorData{}) {
			select {
			case dataChannel <- sensorData:
			default:
				logger.Log.Error("Channel is full, skipping sending value: ", sensorData.Data)
			}
		}

		if nValue != infiniteValue {
			nValue--
		}

		time.Sleep(simulation.TIMEOUT * time.Millisecond)

	}

	logger.Log.Info("Simulation completed.")
	return nil
}

func SimulateForever(dataChannel chan structure.SensorData) {
	logger.Log.Info("Starting sensor simulator loop...")
	for {
		if err := Simulate(infiniteValue, dataChannel); err != nil {
			logger.Log.Error("Error during simulation: ", err)
			os.Exit(1)
		}
	}
}
