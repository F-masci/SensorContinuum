package simulation

import (
	"SensorContinuum/configs/simulation"
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

func Simulate(nValue int, res chan float64) error {

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

		value := generateValue()

		logger.Log.Debug("Sensor reading: ", value.Temperature)

		if res != nil {
			select {
			case res <- value.Temperature:
				logger.Log.Debug("Sent value to channel: ", value.Temperature)
			default:
				logger.Log.Error("Channel is full, skipping sending value: ", value.Temperature)
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

func SimulateForever(res chan float64) {
	logger.Log.Info("Starting sensor simulator loop...")
	for {
		if err := Simulate(infiniteValue, res); err != nil {
			logger.Log.Error("Error during simulation: ", err)
			return
		}
	}
}
