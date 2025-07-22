package intermediate_fog_hub

import (
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
)

func ProcessProximityFogHubData(dataChannel chan structure.SensorData) {
	for data := range dataChannel {
		logger.Log.Info("Received Aggregated Data: ", data)
	}
}
