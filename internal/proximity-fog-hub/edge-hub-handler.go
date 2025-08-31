package proximity_fog_hub

import (
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
)

// ProcessEdgeHubData riceve i dati che arrivano dal Edge Hub tramite MQTT nel canale
// I dati vengono messi nel canale dalla funzione makeSensorDataHandler e, una volta ricevuti,
// li salva nella cache locale.
func ProcessEdgeHubData(dataChannel chan types.SensorData) {
	// si mette in attesa di ricevere i dati
	for data := range dataChannel {
		logger.Log.Info("Filtered data received, sensorId: ", data.SensorID, ", value: ", data.Data)

		// Salva il dato nella cache locale (TimescaleDB)
		// Usiamo un contesto separato per l'operazione sul DB per non bloccare tutto.
		ctx := context.Background()
		if err := storage.InsertSensorData(ctx, data); err != nil {
			// Se il salvataggio fallisce, logghiamo l'errore ma NON CI FERMIAMO.
			// La resilienza impone che l'invio in tempo reale a Kafka abbia la priorità.
			logger.Log.Error("Failure to save data to local cache, sensorId: ", data.SensorID, ", Error: ", err)
		} else {
			logger.Log.Info("Data successfully saved to local cache, sensorId: ", data.SensorID)
		}
	}
}

// ProcessEdgeHubConfiguration riceve i messaggi di configurazione che arrivano dal Edge Hub tramite MQTT nel canale
func ProcessEdgeHubConfiguration(configChannel chan types.ConfigurationMsg) {
	for configMsg := range configChannel {
		if configMsg == (types.ConfigurationMsg{}) {
			logger.Log.Warn("Received empty configuration message, skipping...")
			continue
		}
		logger.Log.Info("Configuration message received: ", configMsg.MsgType)
		// Invia il messaggio di configurazione al Region Hub
		if err := comunication.SendConfigurationMessage(configMsg); err != nil {
			logger.Log.Error("Failure to send configuration message to Region Hub: ", err)
			continue
		}
		logger.Log.Info("Configuration message sent to Region Hub successfully")
		comunication.CleanRetentionConfigurationMessage(configMsg)
		logger.Log.Info("Configuration message cleaned successfully")
	}
}

// ProcessEdgeHubHeartbeat riceve i messaggi di heartbeat che arrivano dal Edge Hub tramite MQTT nel canale
func ProcessEdgeHubHeartbeat(heartbeatChannel chan types.HeartbeatMsg) {
	for heartbeatMsg := range heartbeatChannel {
		if heartbeatMsg == (types.HeartbeatMsg{}) {
			logger.Log.Warn("Received empty heartbeat message, skipping...")
			continue
		}
		logger.Log.Info("Heartbeat message received: ", heartbeatMsg.HubID)
		// Invia il messaggio di heartbeat al Region Hub
		if err := comunication.SendHeartbeatMessage(heartbeatMsg); err != nil {
			logger.Log.Error("Failure to send heartbeat message to Region Hub: ", err)
			continue
		}
		logger.Log.Info("Heartbeat message sent to Region Hub successfully.")
		comunication.CleanRetentionHeartbeatMessage(heartbeatMsg)
		logger.Log.Info("Heartbeat message cleaned successfully.")
	}
}
