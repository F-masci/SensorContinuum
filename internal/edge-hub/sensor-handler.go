package edge_hub

import (
	"SensorContinuum/internal/edge-hub/comunication"
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/internal/edge-hub/processing/aggregation"
	"SensorContinuum/internal/edge-hub/processing/filtering"
	"SensorContinuum/internal/edge-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
	"time"
)

// FilterSensorData orchestra il filtraggio dei dati dei sensori.
func FilterSensorData(dataChannel chan structure.SensorData) {
	storage.InitRedisConnection()
	ctx := context.Background()

	for data := range dataChannel {
		logger.Log.Debug("Processing data for sensor", data.SensorID)

		// 1. Recupera la storia dal Redis
		readings, err := storage.GetSensorHistory(ctx, data.SensorID, environment.HistoryWindowSize)
		if err != nil {
			logger.Log.Error("Error getting sensor history from Redis: ", err)
			continue
		}

		// 2. Controlla se il dato è un outlier BASANDOSI sulla storia attuale (PRIMA di aggiungere il nuovo dato)
		isOutlier := filtering.IsOutlier(data, readings)

		// 3. IN BASE AL RISULTATO del controllo, decidiamo se scartare il dato.
		if isOutlier {
			logger.Log.Warn("Outlier detected and discarded for sensor " + data.SensorID)
			logger.Log.Warn("value outliner detected: ", data.Data)
			logger.Log.Warn("timestamp: ", data.Timestamp)
			continue
		}

		// 4. Se il dato è valido, procedi con l'elaborazione successiva.
		logger.Log.Debug("Data is valid for sensor", data.SensorID)

		// 5. Aggiungi il nuovo dato alla storia su Redis se non è un outlier.
		if err := storage.AddSensorHistory(ctx, data); err != nil {
			logger.Log.Error("Error saving sensor data to Redis: ", err)
			continue
		}
	}
}

// AggregateSensorData aggrega il dato del sensore corrente calcolando la media per il minuto corrente.
func AggregateSensorData(sensorID string) structure.SensorData {
	storage.InitRedisConnection()
	ctx := context.Background()

	now := time.Now().UTC()
	minuteStart := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)
	readings, err := storage.GetSensorHistoryByMinute(ctx, sensorID, minuteStart)
	if err != nil {
		logger.Log.Error("Error getting sensor history from Redis: ", err)
		return structure.SensorData{}
	}

	if len(readings) == 0 {
		logger.Log.Warn("No valid readings found for sensor " + sensorID + " in the current minute")
		return structure.SensorData{}
	}

	// Calcola la media solo sui dati filtrati
	avg := aggregation.AverageCurrentMinute(readings)

	// Prendi BuildingID e FloorID dal primo dato valido
	buildingID := readings[0].BuildingID
	floorID := readings[0].FloorID

	result := structure.SensorData{
		BuildingID: buildingID,
		FloorID:    floorID,
		SensorID:   sensorID,
		Data:       avg,
		Timestamp:  minuteStart.Format(time.RFC3339),
	}
	logger.Log.Info("Average for current minute for sensor "+sensorID+": ", avg)

	// Invio al prossimo Hub
	err = comunication.SendAggregatedData(result)
	if err != nil {
		logger.Log.Error("Error sending average data to Edge Hub - Aggregator Microservice", err)
	}

	return result
}
