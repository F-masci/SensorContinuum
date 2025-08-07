package proximity_fog_hub

import (
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
)

// ProcessEdgeHubData riceve i dati che arrivano dal Edge Hub tramite MQTT nel canale
// (data= <- dataChannel). I dati vengono messi nel canale dalla funzione makeSensorDataHandler
// presente in internal/proximity-fog-hub/communication/mqtt.go. Una volta ricevuti li salva nella cache
//locale (TimescaleDB) e poi li invia a Kafka per l'elaborazione

func ProcessEdgeHubData(dataChannel chan structure.SensorData) {
	// si mette in attesa di ricevere i dati
	for data := range dataChannel {
		logger.Log.Info("Filtered data received, sensorId: ", data.SensorID, ", value:", data.Data)

		// Salva il dato nella cache locale (TimescaleDB)
		// Usiamo un contesto separato per l'operazione sul DB per non bloccare tutto.
		ctx := context.Background()
		if err := storage.InsertSensorData(ctx, data); err != nil {
			// Se il salvataggio fallisce, logghiamo l'errore ma NON CI FERMIAMO.
			// La resilienza impone che l'invio in tempo reale a Kafka abbia la priorità.
			logger.Log.Error("Failure to save data to local cache, sensorId: ", data.SensorID, ", Error: ", err)
		} else {
			logger.Log.Debug("Data successfully saved to local cache, sensorId: ", data.SensorID)
		}

		// una volta fatto il salvataggio nella cache locale, inviamo i dati tramite kafka all intermediate-fog-hub
		logger.Log.Info("Sending data to the Intermediate Fog Hub... sensorId: ", data.SensorID)
		if err := comunication.SendRealTimeData(data); err != nil {
			logger.Log.Error("Failure to send message to Kafka, Error: ", err.Error())
			continue
		} else {
			logger.Log.Info("Message successfully sent to Kafka, sensorId: ", data.SensorID)
		}
	}
}
