package sensor_agent

import (
	"math/rand"
	"time"

	"SensorContinuum/pkg/logger"
)

func Simulate(nValue int, res chan float64) {

	logger.Log.Info("Starting sensor simulator...")

	for nValue > 0 {

		value := rand.Float64() * 100

		logger.Log.Debug("Sensor reading: ", value)

		if res != nil {
			select {
			case res <- value:
				logger.Log.Debug("Sent value to channel: ", value)
			default:
				logger.Log.Error("Channel is full, skipping sending value: ", value)
			}
		}

		nValue--

		time.Sleep(1 * time.Second)

	}
}

func SimulateForever(res chan float64) {
	for {
		Simulate(1000, res)
	}
}
