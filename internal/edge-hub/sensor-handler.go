package edge_hub

import (
	"SensorContinuum/internal/edge-hub/processing"
	"SensorContinuum/internal/edge-hub/processing/filtering"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
)

const (
	// HistoryWindowSize definisce quanti dati storici per sensore conservare per l'analisi.
	HistoryWindowSize = 10
)

// sensorStates mantiene la storia dei dati per ogni sensore.
// La mappa usa l'ID del sensore come chiave.
var sensorStates = make(map[string]*processing.History)

// ProcessSensorData orchestra il processamento dei dati in arrivo.
func ProcessSensorData(dataChannel chan structure.SensorData) {
	for data := range dataChannel {
		logger.Log.Debug("Processing data for sensor", data.SensorID)

		// 1. Ottieni (o crea) la storia per questo specifico sensore
		history, exists := sensorStates[data.SensorID]
		if !exists {
			logger.Log.Info("First time seeing sensor, creating history tracking for", data.SensorID)
			history = processing.NewHistory(HistoryWindowSize)
			sensorStates[data.SensorID] = history
		}

		// 2. Controlla se il dato è un outlier BASANDOSI sulla storia attuale (PRIMA di aggiungere il nuovo dato)
		isOutlier := filtering.IsOutlier(data, history.GetReadings())

		// 3. Aggiungi SEMPRE il nuovo dato alla storia.
		// In questo modo la finestra scorre costantemente, evitando il blocco.
		// La storia conterrà sempre le ultime N misurazioni ricevute.
		history.Add(data)

		// 4. Ora, IN BASE AL RISULTATO del controllo, decidiamo se scartare il dato.
		if isOutlier {
			logger.Log.Warn("Outlier detected and discarded for sensor " + data.SensorID)
			logger.Log.Warn("value outliner detected: ", data.Data)
			logger.Log.Warn("timestamp: ", data.Timestamp)
			// Il dato è un outlier, quindi saltiamo l'invio al Fog Aggregator.
			continue
		}

		// 5. Se il dato è valido, procedi con l'elaborazione successiva.
		logger.Log.Debug("Data is valid for sensor", data.SensorID)

		// 6. (DA IMPLEMENTARE) Invia il dato valido al Fog Aggregator.
		logger.Log.Info("Valid data received from sensor " + data.SensorID)
		logger.Log.Info("value: ", data.Data)
		logger.Log.Info("send data to Fog Aggregator...To do later")
		// Esempio placeholder per il futuro:
		// fog_aggregator.SendData(data)
	}
}
