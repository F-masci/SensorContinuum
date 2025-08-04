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
		// verifichiamo se c'e' un errore nel send dei dati verso l'intermediate-fog-hub
		if err := comunication.SendAggregatedData(data); err != nil {
			// in caso di errore
			logger.Log.Error("Failed to send message to Kafka", "error", err.Error())
			//come strategia decido di continuare a processare i messaggi successivi
			continue
		}

		logger.Log.Info("Successfully sent message to Kafka", "sensor", data.SensorID)
	}
}
