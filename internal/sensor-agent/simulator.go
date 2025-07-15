package sensor_agent

import (
	"math/rand"
	"time"

	"SensorContinuum/pkg/logger"
)

var log = logger.New(getContext())

func Run(nValue int, resChan chan float64) {

	log.Info("Starting sensor simulator...")

	for nValue > 0 {

		value := rand.Float64() * 100

		log.Debug("Sensor reading: ", value)

		if resChan != nil {
			select {
			case resChan <- value:
				log.Debug("Sent value to channel: ", value)
			default:
				log.Error("Channel is full, skipping sending value: ", value)
			}
		} else {
			log.Debug("No channel provided, not sending value: ", value)
		}

		nValue--

		time.Sleep(1 * time.Second)

	}
}

func RunForever() {
	for {
		Run(1000, nil)
	}
}

// TODO: Configurare il contesto del logger dinamicamente
func getContext() logger.Context {
	return logger.Context{
		"service":   "sensor-agent",
		"sensorID":  "sensor-123",
		"requestID": "req-456",
	}
}
