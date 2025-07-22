package proximity_fog_hub

import (
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
)

func ProcessEdgeHubData(dataChannel chan structure.SensorData) {
	for data := range dataChannel {
		logger.Log.Info("Received Filtered Data: ", data)
	}
}
