package proximity_fog_hub

import (
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
)

func ProcessEdgeHubData(dataChannel chan types.SensorData) {
	for data := range dataChannel {
		logger.Log.Info("Filtered data received - sensorId: ", data.SensorID, " - value:", data.Data)

		// --- NUOVA LOGICA: Salva il dato nella cache locale (TimescaleDB) ---
		// Usiamo un contesto separato per l'operazione sul DB per non bloccare tutto.
		ctx := context.Background()
		if err := storage.InsertSensorData(ctx, data); err != nil {
			// Se il salvataggio fallisce, logghiamo l'errore ma NON CI FERMIAMO.
			// La resilienza impone che l'invio in tempo reale a Kafka abbia la priorità.
			logger.Log.Error("Failure to save data to local cache, sensorId: ", data.SensorID, " - error: ", err)
		} else {
			logger.Log.Debug("Data successfully saved to local cache, sensorId: ", data.SensorID)
		}
		// --- FINE NUOVA LOGICA ---

		logger.Log.Info("Sending data to the Intermediate Fog Hub, sensorId: ", data.SensorID)
		if err := comunication.SendAggregatedData(data); err != nil {
			logger.Log.Error("Failure to send message to Kafka, error: ", err.Error())
			continue
		}

		logger.Log.Info("Message successfully sent to Kafka, sensorId: ", data.SensorID)
	}
}

func ProcessEdgeHubConfiguration(configChannel chan types.ConfigurationMsg) {
	for configMsg := range configChannel {
		logger.Log.Info("Configuration message received: ", configMsg.MsgType)
		// Invia il messaggio di configurazione al Region Hub
		if err := comunication.SendConfigurationMessage(configMsg); err != nil {
			logger.Log.Error("Failure to send configuration message to Region Hub, error: ", err)
			continue
		}
		logger.Log.Info("Configuration message sent to Region Hub successfully.")
	}
}
