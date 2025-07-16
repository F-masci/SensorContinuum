package edge_hub

import (
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
)

func ProcessSensorData(dataChannel chan structure.SensorData) {

	for data := range dataChannel {

		logger.Log.Info("Received sensor data", data)
		logger.Log.Debug("Processing sensor data", data)

	}

}
