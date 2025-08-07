package proximity_fog_hub

import (
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
)

func ProcessEdgeHubData(dataChannel chan structure.SensorData) {
	for data := range dataChannel {
		logger.Log.Info("Filtered data received, sensorId: ", data.SensorID, "/n value:", data.Data)

		// --- NUOVA LOGICA: Salva il dato nella cache locale (TimescaleDB) ---
		// Usiamo un contesto separato per l'operazione sul DB per non bloccare tutto.
		ctx := context.Background()
		if err := storage.InsertSensorData(ctx, data); err != nil {
			// Se il salvataggio fallisce, logghiamo l'errore ma NON CI FERMIAMO.
			// La resilienza impone che l'invio in tempo reale a Kafka abbia la priorità.
			logger.Log.Error("Failure to save data to local cache, sensorId: ", data.SensorID, "/n error: ", err)
		} else {
			logger.Log.Debug("Data successfully saved to local cache, sensorId: ", data.SensorID)
		}
		// --- FINE NUOVA LOGICA ---

		logger.Log.Info("Sending data to the Intermediate Fog Hub... sensorId: ", data.SensorID)
		if err := comunication.SendAggregatedData(data); err != nil {
			logger.Log.Error("Failure to send message to Kafka, error: ", err.Error())
			continue
		}

		logger.Log.Info("Message successfully sent to Kafka, sensorId: ", data.SensorID)
	}
}
