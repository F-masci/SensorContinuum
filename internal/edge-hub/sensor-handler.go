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
		logger.Log.Debug("Processing data for sensor: ", data.SensorID)

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
		logger.Log.Debug("Data is valid for sensor: ", data.SensorID)

		// 5. Aggiungi il nuovo dato alla storia su Redis se non è un outlier.
		if err := storage.AddSensorHistory(ctx, data); err != nil {
			logger.Log.Error("Error saving sensor data to Redis: ", err)
			continue
		}
	}
}

// AggregateAllSensorsData esegue l'aggregazione per tutti i sensori presenti in Redis.
func AggregateAllSensorsData() {
	storage.InitRedisConnection()
	ctx := context.Background()

	sensorIDs, err := storage.GetAllSensorIDs(ctx)
	if err != nil {
		logger.Log.Error("Error getting sensor IDs from Redis: ", err)
	}

	var results []structure.SensorData
	now := time.Now().UTC()
	minuteStart := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()-1, 0, 0, time.UTC)

	logger.Log.Info("Starting aggregation for all sensors at ", minuteStart.Format(time.RFC3339))

	for _, sensorID := range sensorIDs {
		readings, err := storage.GetSensorHistoryByMinute(ctx, sensorID, minuteStart)
		if err != nil {
			logger.Log.Error("Error getting sensor history from Redis for sensor ", sensorID, ": ", err)
			continue
		}

		if len(readings) == 0 {
			logger.Log.Warn("No valid readings found for sensor " + sensorID + " in the current minute")
			continue
		}

		avg := aggregation.AverageInMinute(readings, minuteStart)
		buildingID := readings[0].BuildingID
		floorID := readings[0].FloorID

		result := structure.SensorData{
			BuildingID: buildingID,
			FloorID:    floorID,
			SensorID:   sensorID,
			Data:       avg,
			Timestamp:  minuteStart.Format(time.RFC3339),
		}
		logger.Log.Info("Average for minute ", minuteStart.Format(time.RFC3339), " sensor "+sensorID+": ", avg)

		err = comunication.SendAggregatedData(result)
		if err != nil {
			logger.Log.Error("Error sending average data to Proximity Fog Hub", err)
		}

		results = append(results, result)
	}

}

func CleanUnhealthySensors() (unhealthySensors []string) {
	storage.InitRedisConnection()
	ctx := context.Background()

	sensorIDs, err := storage.GetAllSensorIDs(ctx)
	if err != nil {
		logger.Log.Error("Error getting sensor IDs from Redis: ", err)
		return
	}

	for _, sensorID := range sensorIDs {
		readings, err := storage.GetSensorHistory(ctx, sensorID, environment.HistoryWindowSize)
		if err != nil {
			logger.Log.Error("Error getting sensor history from Redis for sensor ", sensorID, ": ", err)
			continue
		}

		if len(readings) == 0 {
			logger.Log.Warn("No readings found for sensor " + sensorID + ". Removing from Redis.")
			if err := storage.RemoveSensorHistory(ctx, sensorID); err != nil {
				logger.Log.Error("Error removing sensor ", sensorID, " from Redis: ", err)
			}
			unhealthySensors = append(unhealthySensors, sensorID)
			continue
		}

		// Trova il timestamp più recente tra le letture
		latestTime := time.Time{}
		for _, reading := range readings {
			t, err := time.Parse(time.RFC3339, reading.Timestamp)
			if err != nil {
				continue
			}
			if t.After(latestTime) {
				latestTime = t
			}
		}

		// Se l'ultima lettura è più vecchia di 5 minuti, elimina la storia
		if time.Since(latestTime) > environment.UnhealthySensorTimeout {
			logger.Log.Warn("Sensor " + sensorID + " inactive for over 5 minutes. Removing from Redis.")
			if err := storage.RemoveSensorHistory(ctx, sensorID); err != nil {
				logger.Log.Error("Error removing sensor ", sensorID, " from Redis: ", err)
			}
			unhealthySensors = append(unhealthySensors, sensorID)
			continue
		}

		logger.Log.Info("Sensor "+sensorID+" for healthy. Readings count: ", len(readings))

	}

	logger.Log.Info("Cleaned up unhealthy sensors. Total sensors checked: ", len(sensorIDs))
	if len(unhealthySensors) > 0 {
		logger.Log.Warn("Unhealthy sensors found: ", unhealthySensors)
	} else {
		logger.Log.Info("No unhealthy sensors found.")
	}
	return unhealthySensors
}

func NotifyUnhealthySensors(unhealthySensors []string) {
	if len(unhealthySensors) == 0 {
		logger.Log.Info("No unhealthy sensors to notify.")
		return
	}

	for _, sensorID := range unhealthySensors {
		logger.Log.Warn("Notifying about unhealthy sensor: ", sensorID)
		// TODO: Implement the actual notification logic
		// err := comunication.NotifyUnhealthySensor(sensorID)
		// if err != nil {
		// 	 logger.Log.Error("Error notifying about unhealthy sensor ", sensorID, ": ", err)
		// }
	}
}
