package intermediate_fog_hub

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/internal/intermediate-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"log"
	"os"
	"time"
)

// setupSensorDbConnection stabilisce la connessione al database dei sensori.
func setupSensorDbConnection() {
	err := storage.SetupSensorDbConnection()
	if err != nil {
		logger.Log.Error("Failed to connect to the sensor database: ", err)
		os.Exit(1)
	}
}

// setupRegionDbConnection stabilisce la connessione al database delle regioni.
func setupRegionDbConnection() {
	err := storage.SetupRegionDbConnection()
	if err != nil {
		logger.Log.Error("Failed to connect to the region database: ", err)
		os.Exit(1)
	}
}

// ProcessRealTimeData gestisce i dati in tempo reale ricevuti dai sensori e li salva in batch.
func ProcessRealTimeData(dataChannel chan types.SensorData) {

	setupSensorDbConnection()
	setupRegionDbConnection()

	var batch = types.NewSensorDataBatch()

	// Timer per la scrittura dei dati in batch
	timer := time.NewTimer(time.Second * time.Duration(environment.SensorDataBatchTimeout))

	for {
		select {
		case data := <-dataChannel:
			logger.Log.Info("Received real-time Data from: ", data.SensorID)

			if batch.Count() >= environment.SensorDataBatchSize {
				logger.Log.Info("Inserting batch data into the database")
				if err := storage.InsertSensorDataBatch(batch); err != nil {
					logger.Log.Error("Failed to insert sensor data batch: ", err)
				}
				logger.Log.Info("Updating last seen for batch sensors")
				if err := storage.UpdateLastSeenBatch(batch); err != nil {
					logger.Log.Error("Failed to update last seen for sensors: ", err)
				}
				logger.Log.Info("Clearing batch after processing")
				batch.Clear()
				timer.Reset(time.Second * 10)
				continue
			}

			logger.Log.Debug("Adding sensor data to batch: ", data.SensorID)
			batch.AddSensorData(data)

		case <-timer.C:
			if batch.Count() > 0 {
				logger.Log.Info("Inserting batch data into the database")
				if err := storage.InsertSensorDataBatch(batch); err != nil {
					logger.Log.Error("Failed to insert sensor data batch: ", err)
				}
				logger.Log.Info("Updating last seen for batch sensors")
				if err := storage.UpdateLastSeenBatch(batch); err != nil {
					logger.Log.Error("Failed to update last seen for sensors: ", err)
				}
				logger.Log.Info("Clearing batch after processing")
				batch.Clear()
			}
			timer.Reset(time.Second * 10)
		}
	}
}

// ProcessStatisticsData gestisce le statistiche aggregate e le salva.
func ProcessStatisticsData(statsChannel chan types.AggregatedStats) {

	setupSensorDbConnection()

	for stats := range statsChannel {

		if stats.Zone == "" {
			logger.Log.Info("Aggregated statistics for macrozone (", stats.Macrozone, ") received: ", stats.Type, " - avg ", stats.Avg)
			if err := storage.InsertMacrozoneStatisticsData(stats); err != nil {
				logger.Log.Error("Failed to insert statistics: ", err)
				os.Exit(-1)
			} else {
				logger.Log.Info("Statistics successfully inserted into the database")
			}
		} else {
			logger.Log.Info("Aggregated statistics for zone (", stats.Macrozone, " - ", stats.Zone, ") received: ", stats.Type, " - avg ", stats.Avg)
			if err := storage.InsertZoneStatisticsData(stats); err != nil {
				logger.Log.Error("Failed to insert statistics: ", err)
				os.Exit(-1)
			} else {
				logger.Log.Info("Statistics successfully inserted into the database")
			}
		}
	}
}

// ProcessProximityFogHubConfiguration gestisce i messaggi di configurazione per il Proximity Fog Hub.
func ProcessProximityFogHubConfiguration(msgChannel chan types.ConfigurationMsg) {

	setupRegionDbConnection()

	for msg := range msgChannel {
		logger.Log.Info("Received configuration message: ", msg)

		if msg.MsgType == types.NewProximityMsgType {
			logger.Log.Debug("Registration of new macrozone hub: ", msg.EdgeMacrozone, ". Hub ID: ", msg.HubID)
			if err := storage.RegisterMacrozoneHub(msg); err != nil {
				log.Fatalf("Registration failed: %v", err)
			}
		} else if msg.MsgType == types.NewEdgeMsgType {
			logger.Log.Debug("Registration of new zone hub: ", msg.EdgeZone, ". Hub ID: ", msg.HubID)
			if err := storage.RegisterZoneHub(msg); err != nil {
				log.Fatalf("Registration failed: %v", err)
			}
		} else if msg.MsgType == types.NewSensorMsgType {
			logger.Log.Debug("Registration of new sensor: ", msg.SensorID, ". Zone Hub ID: ", msg.HubID)
			if err := storage.RegisterSensor(msg); err != nil {
				log.Fatalf("Registration failed: %v", err)
			}
		} else {
			logger.Log.Error("Invalid configuration message types: ", msg.MsgType)
		}

		logger.Log.Info("Configuration message processed: ", msg)

	}
}

// ProcessProximityFogHubHeartbeat gestisce i messaggi di heartbeat per il Proximity Fog Hub.
func ProcessProximityFogHubHeartbeat(heartbeatChannel chan types.HeartbeatMsg) {

	setupRegionDbConnection()

	for heartbeatMsg := range heartbeatChannel {
		logger.Log.Info("Received heartbeat message: ", heartbeatMsg.HubID)

		updateFunc := storage.UpdateLastSeenMacrozoneHub
		if heartbeatMsg.EdgeZone != "" {
			updateFunc = storage.UpdateLastSeenZoneHub
		}

		// Invia il messaggio di heartbeat al Region Hub
		if err := updateFunc(heartbeatMsg); err != nil {
			logger.Log.Error("Failed to update last seen: ", err)
			continue
		}

		logger.Log.Info("Heartbeat message processed successfully for hub: ", heartbeatMsg.HubID)
	}
}
