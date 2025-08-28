package edge_hub

import (
	"SensorContinuum/internal/edge-hub/comunication"
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/internal/edge-hub/processing/aggregation"
	"SensorContinuum/internal/edge-hub/processing/filtering"
	"SensorContinuum/internal/edge-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"time"

	"github.com/google/uuid"
)

var randomInstanceID = uuid.New().String()

// FilterSensorData orchestra il filtraggio dei dati dei sensori.
func FilterSensorData(sensorDataChannel chan types.SensorData) {
	storage.InitRedisConnection()
	ctx := context.Background()

	for data := range sensorDataChannel {
		logger.Log.Info("Processing data for sensor ", data.SensorID, " - value: ", data.Data, ", timestamp: ", data.Timestamp)

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
			logger.Log.Warn("Outlier detected and discarded for sensor ", data.SensorID, " - value: ", data.Data, ", timestamp: ", data.Timestamp)
			continue
		}

		// 4. Se il dato è valido, procedi con l'elaborazione successiva.
		logger.Log.Info("Data is valid for sensor: ", data.SensorID)

		// 5. Aggiungi il nuovo dato alla storia su Redis se non è un outlier.
		if err := storage.AddSensorHistory(ctx, data); err != nil {
			logger.Log.Error("Error saving sensor data to Redis: ", err)
			continue
		}
	}
}

// AggregateAllSensorsData esegue l'aggregazione per tutti i sensori presenti in Redis.
func AggregateAllSensorsData(filteredDataChannel chan types.SensorData) {
	storage.InitRedisConnection()
	ctx := context.Background()

	// Prova a ottenere il lock di leader, solo il leader esegue l'aggregazione
	var instanceID = environment.HubID + "-" + randomInstanceID
	isLeader, err := storage.TryOrRenewLeader(ctx, instanceID)
	if err != nil {
		logger.Log.Error("Error acquiring leader lock: ", err)
		return
	}
	if !isLeader {
		logger.Log.Info("Not the leader, skipping aggregation")
		return
	}

	// Recupera tutti gli ID dei sensori
	sensorIDs, err := storage.GetAllSensorIDs(ctx)
	if err != nil {
		logger.Log.Error("Error getting sensor IDs from Redis: ", err)
	}

	var results []types.SensorData

	// Calcola l'inizio del minuto corrente considerando l'offset
	// Ad esempio, se ora è 12:34:45 e l'offset è -2min,
	// minuteStart sarà 12:32:00 e aggrego i dati tra 12:32:00 e 12:32:59
	now := time.Now().UTC()
	minuteStart := now.Add(environment.AggregationFetchOffset).Truncate(time.Minute)

	logger.Log.Info("Starting aggregation for all sensors at ", minuteStart.Format(time.RFC3339))

	for _, sensorID := range sensorIDs {

		// Recupera le letture del sensore per il minuto specificato
		readings, err := storage.GetSensorHistoryByMinute(ctx, sensorID, minuteStart)
		if err != nil {
			logger.Log.Error("Error getting sensor history from Redis for sensor ", sensorID, ": ", err)
			continue
		}

		// Se non ci sono letture valide, salta questo sensore
		if len(readings) == 0 {
			logger.Log.Warn("No valid readings found for sensor " + sensorID + " in the current minute")
			continue
		}

		// Calcola la media delle letture per il minuto specificato
		avg := aggregation.AverageInMinute(readings, minuteStart)
		edgeMacrozone := readings[0].EdgeMacrozone
		edgeZone := readings[0].EdgeZone
		sensorType := readings[0].Type

		// Crea il risultato dell'aggregazione
		result := types.SensorData{
			EdgeMacrozone: edgeMacrozone,
			EdgeZone:      edgeZone,
			SensorID:      sensorID,
			Data:          avg,
			Timestamp:     minuteStart.Unix(),
			Type:          sensorType,
		}
		logger.Log.Info("Average for minute ", minuteStart.Format(time.RFC3339), " sensor "+sensorID+": ", avg)

		// Invia il risultato al canale di dati filtrati
		select {
		// invia il risultato al canale filteredDataChannel
		case filteredDataChannel <- result:
			logger.Log.Debug("Sent aggregated data for sensor: ", sensorID)
		default:
			logger.Log.Warn("Filtered data channel is full, discarding aggregated data for sensor: ", sensorID)
		}

		// Aggiungi il risultato all'elenco dei risultati
		// per eventuali operazioni successive
		results = append(results, result)
	}

}

// CleanUnhealthySensors rimuove i sensori che non comunicano da troppo tempo.
func CleanUnhealthySensors() (unhealthySensors []string, removedSensors []string) {
	storage.InitRedisConnection()
	ctx := context.Background()

	// Recupera tutti gli ID dei sensori
	sensorIDs, err := storage.GetAllSensorIDs(ctx)
	if err != nil {
		logger.Log.Error("Error getting sensor IDs from Redis: ", err)
		return
	}

	// Controlla lo stato di ogni sensore
	for _, sensorID := range sensorIDs {
		readings, err := storage.GetSensorHistory(ctx, sensorID, environment.HistoryWindowSize)
		if err != nil {
			logger.Log.Error("Error getting sensor history from Redis for sensor ", sensorID, ": ", err)
			continue
		}

		// Trova il timestamp più recente tra le letture
		latestTime := time.Time{}
		for _, reading := range readings {
			t := time.Unix(reading.Timestamp, 0).UTC()
			if t.After(latestTime) {
				latestTime = t
			}
		}

		// Caso: sensore non comunica da troppo tempo (UnhealthySensorTimeout) -> rimuovo la storia del sensore
		if time.Since(latestTime) > environment.UnhealthySensorTimeout {
			logger.Log.Warn("Sensor " + sensorID + " inactive for too times. Removing history from Redis.")
			if err := storage.RemoveSensorHistory(ctx, sensorID); err != nil {
				logger.Log.Error("Error removing sensor ", sensorID, " from Redis: ", err)
			}
			unhealthySensors = append(unhealthySensors, sensorID)
			continue
		}

		// Caso: sensore non comunica da troppo tempo (RegistrationSensorTimeout) -> rimuovo il sensore
		if time.Since(latestTime) > environment.RegistrationSensorTimeout {
			logger.Log.Warn("Sensor " + sensorID + " inactive for too times. Removing from Redis.")
			if err := storage.RemoveSensor(ctx, sensorID); err != nil {
				logger.Log.Error("Error removing sensor ", sensorID, " from Redis: ", err)
			}
			removedSensors = append(removedSensors, sensorID)
			continue
		}

		logger.Log.Info("Checked sensor "+sensorID+" for healthy. Readings count: ", len(readings))
	}

	logger.Log.Info("Cleaned up unhealthy sensors. Total sensors checked: ", len(sensorIDs))
	if len(unhealthySensors) > 0 {
		logger.Log.Warn("Unhealthy sensors found: ", unhealthySensors)
	} else {
		logger.Log.Info("No unhealthy sensors found.")
	}
	if len(removedSensors) > 0 {
		logger.Log.Warn("Removed sensors: ", removedSensors)
	} else {
		logger.Log.Info("No sensors removed.")
	}
	return unhealthySensors, removedSensors
}

// ProcessSensorConfigurationMessages gestisce i messaggi di configurazione dei sensori.
func ProcessSensorConfigurationMessages(sensorConfigurationMessageChannel, hubConfigurationMessageChannel chan types.ConfigurationMsg) {
	storage.InitRedisConnection()
	ctx := context.Background()

	for configMsg := range sensorConfigurationMessageChannel {
		logger.Log.Info("Processing configuration message for sensor: ", configMsg.SensorID)

		// Esegui le operazioni necessarie in base al tipo di configurazione
		switch configMsg.MsgType {
		case types.NewSensorMsgType:

			// Aggiungi il sensore al database se non esiste già
			sensor := types.Sensor{
				Id:            configMsg.SensorID,
				ZoneName:      configMsg.EdgeZone,
				MacrozoneName: configMsg.EdgeMacrozone,
				Type:          configMsg.SensorType,
				Reference:     configMsg.SensorReference,
			}

			// Aggiungo il sensore solo se non esiste già
			// Se il sensore è nuovo, inoltro il messaggio di configurazione all'hub
			// Altrimenti, scarto il messaggio di configurazione
			// In entrambi i casi, pulisco il messaggio di configurazione
			// per evitare di ritrasmetterlo in futuro
			// (ad esempio, se il sensore si riavvia e invia di nuovo il messaggio di configurazione)
			if exists, err := storage.AddSensor(ctx, sensor); err != nil {
				logger.Log.Error("Error adding sensor configuration: ", err)
			} else if !exists {
				logger.Log.Info("Sensor configuration added for sensor: ", configMsg.SensorID)
				hubConfigurationMessageChannel <- configMsg
			} else {
				logger.Log.Info("Sensor configuration already exists for sensor: ", configMsg.SensorID)
				comunication.CleanRetentionConfigurationMessage(configMsg)
			}
		default:
			logger.Log.Warn("Unknown configuration types for sensor: ", configMsg.SensorID)
		}
	}
}

// NotifyUnhealthySensors invia notifiche per i sensori non sani.
func NotifyUnhealthySensors(unhealthySensors []string) {
	if len(unhealthySensors) == 0 {
		logger.Log.Info("No unhealthy sensors to notify.")
		return
	}

	for _, sensorID := range unhealthySensors {
		logger.Log.Warn("Notifying about unhealthy sensor: ", sensorID)
		// TODO: Implement the actual notification logic
	}
}

// NotifyRemovedSensors invia notifiche per i sensori rimossi.
func NotifyRemovedSensors(removedSensors []string) {
	if len(removedSensors) == 0 {
		logger.Log.Info("No removed sensors to notify.")
		return
	}

	for _, sensorID := range removedSensors {
		logger.Log.Warn("Notifying about removed sensor: ", sensorID)
		// TODO: Implement the actual notification logic
	}
}
