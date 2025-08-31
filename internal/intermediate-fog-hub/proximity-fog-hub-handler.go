package intermediate_fog_hub

import (
	"SensorContinuum/internal/intermediate-fog-hub/comunication"
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/internal/intermediate-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
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
func ProcessRealTimeData(dataChannel chan types.SensorData, kafkaPauseSignal *utils.PauseSignal) {

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
			// Manda un segnale per mettere in pausa il consumer Kafka
			kafkaPauseSignal.Send(true)
			if err := storage.InsertSensorDataBatch(b); err != nil {
				logger.Log.Error("Failed to insert sensor data batch: ", err)
				os.Exit(1)
			}
			if err := storage.UpdateSensorLastSeenBatch(b); err != nil {
				logger.Log.Error("Failed to update last seen for sensors: ", err)
				os.Exit(1)
			}
			// Se tutto è andato a buon fine, esegui il commit
			// dei messaggi Kafka
			err := comunication.CommitSensorDataBatchMessages(b.GetKafkaMessages())
			if err != nil {
				logger.Log.Error("Failed to commit Kafka messages for sensor data batch: ", err)
				os.Exit(1)
			}
			// Manda un segnale per riavviare il consumer Kafka
			kafkaPauseSignal.Send(false)
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

	logger.Log.Warn("Data channel closed, stopping real-time data processing")
	os.Exit(1)
}

// ProcessStatisticsData gestisce le statistiche aggregate e le salva.
func ProcessStatisticsData(statsChannel chan types.AggregatedStats, kafkaZonePauseSignal, kafkaMacrozonePauseSignal *utils.PauseSignal) {

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
			// Manda un segnale per mettere in pausa il consumer Kafka
			kafkaMacrozonePauseSignal.Send(true)
			if err := storage.InsertMacrozoneStatisticsDataBatch(b); err != nil {
				logger.Log.Error("Failed to insert macrozone aggregated stats batch: ", err)
				os.Exit(1)
			}
			// Se tutto è andato a buon fine, esegui il commit
			// dei messaggi Kafka
			err := comunication.CommitStatisticsDataBatchMessages(b.GetKafkaMessages())
			if err != nil {
				logger.Log.Error("Failed to commit Kafka messages for aggregated stats batch: ", err)
				os.Exit(1)
			}
			// Manda un segnale per riavviare il consumer Kafka
			kafkaMacrozonePauseSignal.Send(false)
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
			// Manda un segnale per mettere in pausa il consumer Kafka
			kafkaZonePauseSignal.Send(true)
			if err := storage.InsertZoneStatisticsDataBatch(b); err != nil {
				logger.Log.Error("Failed to insert zone aggregated stats batch: ", err)
				os.Exit(1)
			}
			// Se tutto è andato a buon fine, esegui il commit
			// dei messaggi Kafka
			err := comunication.CommitStatisticsDataBatchMessages(b.GetKafkaMessages())
			if err != nil {
				logger.Log.Error("Failed to commit Kafka messages for aggregated stats batch: ", err)
				os.Exit(1)
			}
			// Manda un segnale per riavviare il consumer Kafka
			kafkaZonePauseSignal.Send(false)
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

	logger.Log.Warn("Statistics channel closed, stopping statistics data processing")
	os.Exit(1)
}

// ProcessProximityFogHubConfiguration gestisce i messaggi di configurazione per il Proximity Fog Hub.
func ProcessProximityFogHubConfiguration(msgChannel chan types.ConfigurationMsg, kafkaPauseSignal *utils.PauseSignal) {

	// Connessione ai databases
	setupRegionDbConnection()

	// Batch per i messaggi di configurazione
	batch, err := types.NewConfigurationMsgBatch(
		environment.ConfigurationMessageBatchSize,
		time.Duration(environment.ConfigurationMessageBatchTimeout)*time.Second,
		// Funzione di salvataggio dei messaggi
		// Viene chiamata quando il batch è pieno o scade il timeout
		// Salva i messaggi nel database
		func(b *types.ConfigurationMsgBatch) error {
			// Manda un segnale per mettere in pausa il consumer Kafka
			kafkaPauseSignal.Send(true)
			err := storage.RegisterDevicesFromBatch(b)
			if err != nil {
				logger.Log.Error("Failed to register devices from configuration message batch: ", err)
				os.Exit(1)
			}
			// Se tutto è andato a buon fine, esegui il commit
			// dei messaggi Kafka
			err = comunication.CommitConfigurationBatchMessages(b.GetKafkaMessages())
			if err != nil {
				logger.Log.Error("Failed to commit Kafka messages for configuration message batch: ", err)
				os.Exit(1)
			}
			// Manda un segnale per riavviare il consumer Kafka
			kafkaPauseSignal.Send(false)
			return nil
		})
	if err != nil {
		logger.Log.Error("Failed to create configuration message batch: ", err)
		os.Exit(1)
	}

	for msg := range msgChannel {
		logger.Log.Info("Received configuration message: ", msg)
		batch.Add(msg)
	}

	logger.Log.Warn("Configuration channel closed, stopping configuration message processing")
	os.Exit(1)
}

// ProcessProximityFogHubHeartbeat gestisce i messaggi di heartbeat per il Proximity Fog Hub.
func ProcessProximityFogHubHeartbeat(heartbeatChannel chan types.HeartbeatMsg, kafkaPauseSignal *utils.PauseSignal) {

	// Connessione ai databases
	setupRegionDbConnection()

	batch, err := types.NewHeartbeatMsgBatch(
		environment.HeartbeatMessageBatchSize,
		time.Duration(environment.HeartbeatMessageBatchTimeout)*time.Second,
		// Funzione di salvataggio dei messaggi
		// Viene chiamata quando il batch è pieno o scade il timeout
		// Salva i messaggi nel database
		func(b *types.HeartbeatMsgBatch) error {
			// Manda un segnale per mettere in pausa il consumer Kafka
			kafkaPauseSignal.Send(true)
			err := storage.UpdateHubLastSeen(b)
			if err != nil {
				logger.Log.Error("Failed to update last seen from heartbeat message batch: ", err)
				os.Exit(1)
			}
			// Se tutto è andato a buon fine, esegui il commit
			// dei messaggi Kafka
			err = comunication.CommitHeartbeatBatchMessages(b.GetKafkaMessages())
			if err != nil {
				logger.Log.Error("Failed to commit Kafka messages for heartbeat message batch: ", err)
				os.Exit(1)
			}
			// Manda un segnale per riavviare il consumer Kafka
			kafkaPauseSignal.Send(false)
			return nil
		})
	if err != nil {
		logger.Log.Error("Failed to create heartbeat message batch: ", err)
		os.Exit(1)
	}

	for heartbeatMsg := range heartbeatChannel {
		logger.Log.Info("Received heartbeat message: ", heartbeatMsg.HubID)
		batch.Add(heartbeatMsg)
	}

	logger.Log.Warn("Heartbeat channel closed, stopping heartbeat message processing")
	os.Exit(1)
}
