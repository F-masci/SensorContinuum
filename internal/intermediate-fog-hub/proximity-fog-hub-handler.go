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

	// Connessione ai databases
	setupSensorDbConnection()
	setupRegionDbConnection()

	// Batch per i dati dei sensori
	batch, err := types.NewSensorDataBatch(
		environment.SensorDataBatchSize,
		time.Duration(environment.SensorDataBatchTimeout)*time.Second,
		// Funzione di salvataggio dei dati
		// Viene chiamata quando il batch è pieno o scade il timeout
		// Salva i dati nel database e aggiorna il last seen dei sensori
		func(b *types.SensorDataBatch) error {
			if err := storage.InsertSensorDataBatch(*b); err != nil {
				logger.Log.Error("Failed to insert sensor data batch: ", err)
				return err
			}
			if err := storage.UpdateLastSeenBatch(*b); err != nil {
				logger.Log.Error("Failed to update last seen for sensors: ", err)
				return err
			}
			return nil
		})
	if err != nil {
		logger.Log.Error("Failed to create sensor data batch: ", err)
		os.Exit(1)
	}

	for data := range dataChannel {
		logger.Log.Info("Real-time sensor data received: ", data)
		batch.AddSensorData(data)
	}
}

// ProcessStatisticsData gestisce le statistiche aggregate e le salva.
func ProcessStatisticsData(statsChannel chan types.AggregatedStats) {

	// Connessione al database dei sensori
	setupSensorDbConnection()

	// Batch per le statistiche aggregate a livello di macrozona
	macrozoneBatch, err := types.NewAggregatedStatsBatch(
		environment.AggregatedDataBatchSize,
		time.Duration(environment.AggregatedDataBatchTimeout)*time.Second,
		// Funzione di salvataggio delle statistiche
		// Viene chiamata quando il batch è pieno o scade il timeout
		// Salva le statistiche nel database
		func(b *types.AggregatedStatsBatch) error {
			if err := storage.InsertMacrozoneStatisticsDataBatch(*b); err != nil {
				logger.Log.Error("Failed to insert macrozone aggregated stats batch: ", err)
				return err
			}
			return nil
		})
	if err != nil {
		logger.Log.Error("Failed to create aggregated stats batch: ", err)
		os.Exit(1)
	}

	// Batch per le statistiche aggregate a livello di zona
	zoneBatch, err := types.NewAggregatedStatsBatch(
		environment.AggregatedDataBatchSize,
		time.Duration(environment.AggregatedDataBatchTimeout)*time.Second,
		// Funzione di salvataggio delle statistiche
		// Viene chiamata quando il batch è pieno o scade il timeout
		// Salva le statistiche nel database
		func(b *types.AggregatedStatsBatch) error {
			if err := storage.InsertZoneStatisticsDataBatch(*b); err != nil {
				logger.Log.Error("Failed to insert zone aggregated stats batch: ", err)
				return err
			}
			return nil
		})
	if err != nil {
		logger.Log.Error("Failed to create aggregated stats batch: ", err)
		os.Exit(1)
	}

	for stats := range statsChannel {
		logger.Log.Info("Aggregated stats received: ", stats)
		if stats.Zone != "" {
			zoneBatch.AddAggregatedStats(stats)
		} else if stats.Macrozone != "" && stats.Zone == "" {
			macrozoneBatch.AddAggregatedStats(stats)
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
