package proximity_fog_hub

import (
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
)

func ProcessEdgeHubData(dataChannel chan structure.SensorData) {
	for data := range dataChannel {
		logger.Log.Info("Received Filtered Data: ", data)
		logger.Log.Info("Send data to Fog Intermediate Hub: ", data)
		comunication.SendAggregatedData(data)
	}
}
