package comunication

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

// kafkaRealTimeDataReader è il lettore Kafka per i dati in tempo reale.
var kafkaRealTimeDataReader *kafka.Reader = nil

// kafkaStatisticsDataReader è il lettore Kafka per i dati statistici aggregati.
var kafkaStatisticsDataReader *kafka.Reader = nil

// kafkaConfigurationReader è il lettore Kafka per i messaggi di configurazione.
var kafkaConfigurationReader *kafka.Reader = nil

// kafkaHeartbeatReader è il lettore Kafka per i messaggi di heartbeat.
var kafkaHeartbeatReader *kafka.Reader = nil

// connectRealTimeData si connette a Kafka per leggere i dati in tempo reale.
func connectRealTimeData() {

	// Se la connessione è già stabilita, non fare nulla
	if kafkaRealTimeDataReader != nil {
		return // already connected
	}

	logger.Log.Debug("Connecting to Kafka topic: ", environment.ProximityDataTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configura il lettore Kafka per i dati in tempo reale
	kafkaRealTimeDataReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.ProximityDataTopic,
		GroupID: environment.KafkaGroupId,
	})
	logger.Log.Info("Connected to Kafka topic: ", environment.ProximityDataTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)
}

// connectProximityConfiguration si connette a Kafka per leggere i messaggi di configurazione.
func connectProximityConfiguration() {

	// Se la connessione è già stabilita, non fare nulla
	if kafkaConfigurationReader != nil {
		return // already connected
	}

	logger.Log.Debug("Connecting to Kafka topic: ", environment.ProximityConfigurationTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configure the Kafka reader
	kafkaConfigurationReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.ProximityConfigurationTopic,
		GroupID: environment.KafkaGroupId,
	})
	logger.Log.Info("Connected to Kafka topic: ", environment.ProximityConfigurationTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)
}

// connectProximityHeartbeat si connette a Kafka per leggere i messaggi di heartbeat.
func connectProximityHeartbeat() {

	// Se la connessione è già stabilita, non fare nulla
	if kafkaHeartbeatReader != nil {
		return
	}

	logger.Log.Debug("Connecting to Kafka topic: ", environment.ProximityHeartbeatTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configure the Kafka reader
	kafkaHeartbeatReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.ProximityHeartbeatTopic,
		GroupID: environment.KafkaGroupId,
	})
	logger.Log.Info("Connected to Kafka topic: ", environment.ProximityHeartbeatTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)
}

// connectStatisticsData si connette a Kafka per leggere i dati statistici aggregati.
func connectStatisticsData() {

	// Se la connessione è già stabilita, non fare nulla
	if kafkaStatisticsDataReader != nil {
		return // already connected
	}

	logger.Log.Debug("Connecting to Kafka topic: ", environment.AggregatedStatsTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)

	// Configure the Kafka reader
	kafkaStatisticsDataReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{environment.KafkaBroker + ":" + environment.KafkaPort},
		Topic:   environment.AggregatedStatsTopic,
		GroupID: environment.KafkaGroupId,
	})
	logger.Log.Info("Connected to Kafka topic: ", environment.AggregatedStatsTopic, " at ", environment.KafkaBroker+":"+environment.KafkaPort)
}

// PullRealTimeData si occupa di leggere i dati dei sensori in tempo reale.
func PullRealTimeData(dataChannel chan types.SensorData, pauseSignal *utils.PauseSignal) error {

	// Connessione a Kafka se non è già stabilita
	connectRealTimeData()
	ctx := context.Background()
	paused := false

	for {
		select {
		case p := <-pauseSignal.Chan():
			paused = p
			if paused {
				logger.Log.Info("Pausing real-time data consumption from Kafka.")
			} else {
				logger.Log.Info("Resuming real-time data consumption from Kafka.")
			}
		default:

			// Se siamo in pausa, aspetta finché non viene tolta la pausa
			// o finché il contesto non viene cancellato
			// In questo modo evitiamo di leggere messaggi da Kafka
			// quando non siamo pronti a processarli
			if paused {
				logger.Log.Info("Paused real-time data consumption. Waiting for resume...")

				select {
				case p := <-pauseSignal.Chan():
					paused = p
					if paused {
						logger.Log.Info("Real-time data consumption still paused.")
						continue
					} else {
						logger.Log.Info("Resumed real-time data consumption from Kafka.")
					}
				case <-ctx.Done():
					logger.Log.Info("Context canceled while paused. Stopping consumer.")
					return ctx.Err()
				}
			}

			// Legge il messaggio dal topic Kafka
			m, err := kafkaRealTimeDataReader.FetchMessage(ctx)
			if err != nil {
				return err
			}
			logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))

			// Converte il messaggio in un oggetto SensorData
			var data types.SensorData
			data, err = types.CreateSensorDataFromKafka(m)
			if err != nil {
				logger.Log.Error("Error unmarshalling Sensor Data: ", err)
				continue
			}

			// Prova a inviare il dato al canale
			// Se il canale è pieno, ritenta per un numero massimo di volte
			// con un ritardo tra i tentativi
			// Se non riesce a inviare il dato, lo scarta e logga un errore
			// per evitare di bloccare il lettore Kafka
			sent := false
			for attempt := 0; attempt < environment.KafkaMaxAttempts && !sent; attempt++ {
				select {
				case dataChannel <- data:
					// Inviato con successo
					logger.Log.Debug("Sensor data sent to channel: ", data)
					sent = true
				default:
					// Canale pieno, logghiamo un avviso e attendiamo l'elaborazione
					logger.Log.Warn("Data channel is full. Attempt(s) ", attempt+1, " of ", environment.KafkaMaxAttempts)
					if attempt == environment.KafkaMaxAttempts-1 {
						logger.Log.Error("Max attempts reached, discarding sensor data: ", data)
						break
					}
					// Attende prima di ritentare
					time.Sleep(time.Duration(environment.KafkaAttemptDelay) * time.Millisecond)
				}
			}
		}
	}
}

// CommitSensorDataBatchMessages esegue il commit degli offset dei messaggi Kafka in un batch di dati sensori.
func CommitSensorDataBatchMessages(messages []kafka.Message) error {
	// Se il lettore Kafka non è inizializzato, non fare nulla
	if kafkaRealTimeDataReader == nil {
		return nil
	}

	if len(messages) == 0 {
		return nil
	}

	// Esegue il commit dei messaggi
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(environment.KafkaCommitTimeout)*time.Second)
	defer cancel()

	err := kafkaRealTimeDataReader.CommitMessages(ctx, messages...)
	if err != nil {
		logger.Log.Error("Failed to commit Kafka messages: ", err)
		return err
	}

	logger.Log.Debug("Committed Kafka ", len(messages), " messages")
	return nil
}

// PullStatisticsData si occupa di leggere i dati statistici aggregati.
func PullStatisticsData(statsChannel chan types.AggregatedStats, zonePauseSignal, macrozonePauseSignal *utils.PauseSignal) error {

	// Connessione a Kafka se non è già stabilita
	connectStatisticsData()
	ctx := context.Background()
	zonePaused := false
	macrozonePaused := false

	for {
		select {
		case p := <-zonePauseSignal.Chan():
			zonePaused = p
			if zonePaused {
				logger.Log.Info("Pausing statistics data consumption from Kafka.")
			} else {
				logger.Log.Info("Resuming statistics data consumption from Kafka.")
			}
		case p := <-macrozonePauseSignal.Chan():
			macrozonePaused = p
			if macrozonePaused {
				logger.Log.Info("Pausing statistics data consumption from Kafka.")
			} else {
				logger.Log.Info("Resuming statistics data consumption from Kafka.")
			}
		default:

			// Se siamo in pausa, aspetta finché non viene tolta la pausa
			// o finché il contesto non viene cancellato
			// In questo modo evitiamo di leggere messaggi da Kafka
			// quando non siamo pronti a processarli
			if zonePaused || macrozonePaused {
				logger.Log.Info("Paused statistics data consumption. Waiting for resume...")

				select {
				case p := <-zonePauseSignal.Chan():
					zonePaused = p
					if !zonePaused && !macrozonePaused {
						logger.Log.Info("Resuming statistics data consumption from Kafka.")
					}
				case p := <-macrozonePauseSignal.Chan():
					macrozonePaused = p
					if !zonePaused && !macrozonePaused {
						logger.Log.Info("Resuming statistics data consumption from Kafka.")
					}
				case <-ctx.Done():
					logger.Log.Info("Context canceled while paused. Stopping consumer.")
					return ctx.Err()
				}

				// se ancora in pausa, salta al prossimo giro
				if zonePaused || macrozonePaused {
					logger.Log.Info("Statistics data consumption still paused.")
					continue
				}
			}

			// Legge il messaggio dal topic Kafka
			m, err := kafkaStatisticsDataReader.FetchMessage(ctx)
			if err != nil {
				logger.Log.Error("Error reading message: ", err.Error())
				return err
			}
			logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))

			// Converte il messaggio in un oggetto AggregatedStats
			var stats types.AggregatedStats
			stats, err = types.CreateAggregatedStatsFromKafka(m)
			if err != nil {
				logger.Log.Error("Error unmarshalling Aggregated Stats", "error", err)
				continue
			}

			// Prova a inviare le statistiche al canale
			// Se il canale è pieno, ritenta per un numero massimo di volte
			// con un ritardo tra i tentativi
			// Se non riesce a inviare le statistiche, le scarta e logga un errore
			// per evitare di bloccare il lettore Kafka
			sent := false
			for attempt := 0; attempt < environment.KafkaMaxAttempts && !sent; attempt++ {
				select {
				case statsChannel <- stats:
					// Inviato con successo
					logger.Log.Debug("Aggregated stats sent to channel: ", stats)
					sent = true
				default:
					// Canale pieno, logghiamo un avviso e attendiamo l'elaborazione
					logger.Log.Warn("Stats channel is full. Attempt(s) ", attempt+1, " of ", environment.KafkaMaxAttempts)
					if attempt == environment.KafkaMaxAttempts-1 {
						logger.Log.Error("Max attempts reached, discarding aggregated stats: ", stats)
						break
					}
					// Attende prima di ritentare
					time.Sleep(time.Duration(environment.KafkaAttemptDelay) * time.Millisecond)
				}
			}
		}
	}
}

// CommitStatisticsDataBatchMessages esegue il commit degli offset dei messaggi Kafka in un batch di dati statistici.
func CommitStatisticsDataBatchMessages(messages []kafka.Message) error {
	// Se il lettore Kafka non è inizializzato, non fare nulla
	if kafkaStatisticsDataReader == nil {
		return nil
	}

	if len(messages) == 0 {
		return nil
	}

	// Esegue il commit dei messaggi
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(environment.KafkaCommitTimeout)*time.Second)
	defer cancel()

	err := kafkaStatisticsDataReader.CommitMessages(ctx, messages...)
	if err != nil {
		logger.Log.Error("Failed to commit Kafka messages: ", err)
		return err
	}

	logger.Log.Debug("Committed Kafka ", len(messages), " messages")
	return nil
}

// PullConfigurationMessage si occupa di leggere i messaggi di configurazione.
func PullConfigurationMessage(msgChannel chan types.ConfigurationMsg, pauseSignal *utils.PauseSignal) error {

	// Connessione a Kafka se non è già stabilita
	connectProximityConfiguration()
	ctx := context.Background()
	paused := false

	for {
		select {
		case p := <-pauseSignal.Chan():
			paused = p
			if paused {
				logger.Log.Info("Pausing configuration message consumption from Kafka.")
			} else {
				logger.Log.Info("Resuming configuration message consumption from Kafka.")
			}
		default:

			// Se siamo in pausa, aspetta finché non viene tolta la pausa
			// o finché il contesto non viene cancellato
			// In questo modo evitiamo di leggere messaggi da Kafka
			// quando non siamo pronti a processarli
			if paused {
				logger.Log.Info("Paused configuration message consumption. Waiting for resume...")

				select {
				case p := <-pauseSignal.Chan():
					paused = p
					if paused {
						logger.Log.Info("Configuration message consumption still paused.")
						continue
					} else {
						logger.Log.Info("Resumed configuration message consumption from Kafka.")
					}
				case <-ctx.Done():
					logger.Log.Info("Context canceled while paused. Stopping consumer.")
					return ctx.Err()
				}
			}

			// Legge il messaggio dal topic Kafka
			m, err := kafkaConfigurationReader.FetchMessage(ctx)
			if err != nil {
				logger.Log.Error("Error reading message: ", err.Error())
				return err
			}
			logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))

			// Converte il messaggio in un oggetto ConfigurationMsg
			var confMsg types.ConfigurationMsg
			confMsg, err = types.CreateConfigurationMsgFromKafka(m)
			if err != nil {
				logger.Log.Error("Error unmarshalling Configuration Message: ", err.Error())
				continue
			}

			// Prova a inviare il messaggio al canale
			// Se il canale è pieno, ritenta per un numero massimo di volte
			// con un ritardo tra i tentativi
			// Se non riesce a inviare il messaggio, lo scarta e logga un errore
			// per evitare di bloccare il lettore Kafka
			sent := false
			for attempt := 0; attempt < environment.KafkaMaxAttempts && !sent; attempt++ {
				select {
				case msgChannel <- confMsg:
					// Inviato con successo
					logger.Log.Debug("Configuration message sent to channel: ", confMsg)
					sent = true
				default:
					// Canale pieno, logghiamo un avviso e attendiamo l'elaborazione
					logger.Log.Warn("Configuration message channel is full. Attempt(s) ", attempt+1, " of ", environment.KafkaMaxAttempts)
					if attempt == environment.KafkaMaxAttempts-1 {
						logger.Log.Error("Max attempts reached, discarding configuration message: ", confMsg)
						break
					}
					// Attende prima di ritentare
					time.Sleep(time.Duration(environment.KafkaAttemptDelay) * time.Millisecond)
				}
			}
		}
	}
}

// CommitConfigurationBatchMessages esegue il commit degli offset dei messaggi Kafka in un batch di messaggi di configurazione.
func CommitConfigurationBatchMessages(messages []kafka.Message) error {
	// Se il lettore Kafka non è inizializzato, non fare nulla
	if kafkaConfigurationReader == nil {
		return nil
	}

	if len(messages) == 0 {
		return nil
	}

	// Esegue il commit dei messaggi
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(environment.KafkaCommitTimeout)*time.Second)
	defer cancel()

	err := kafkaConfigurationReader.CommitMessages(ctx, messages...)
	if err != nil {
		logger.Log.Error("Failed to commit Kafka messages: ", err)
		return err
	}

	logger.Log.Debug("Committed Kafka ", len(messages), " messages")
	return nil
}

// PullHeartbeatMessage si occupa di leggere i messaggi di heartbeat.
func PullHeartbeatMessage(heartbeatChannel chan types.HeartbeatMsg, pauseSignal *utils.PauseSignal) error {

	// Connessione a Kafka se non è già stabilita
	connectProximityHeartbeat()
	ctx := context.Background()
	paused := false

	for {
		select {
		case p := <-pauseSignal.Chan():
			paused = p
			if paused {
				logger.Log.Info("Pausing heartbeat message consumption from Kafka.")
			} else {
				logger.Log.Info("Resuming heartbeat message consumption from Kafka.")
			}
		default:

			// Se siamo in pausa, aspetta finché non viene tolta la pausa
			// o finché il contesto non viene cancellato
			// In questo modo evitiamo di leggere messaggi da Kafka
			// quando non siamo pronti a processarli
			if paused {
				logger.Log.Info("Paused heartbeat message consumption. Waiting for resume...")

				select {
				case p := <-pauseSignal.Chan():
					paused = p
					if paused {
						logger.Log.Info("Heartbeat message consumption still paused.")
						continue
					} else {
						logger.Log.Info("Resumed heartbeat message consumption from Kafka.")
					}
				case <-ctx.Done():
					logger.Log.Info("Context canceled while paused. Stopping consumer.")
					return ctx.Err()
				}
			}

			// Legge il messaggio dal topic Kafka
			m, err := kafkaHeartbeatReader.FetchMessage(ctx)
			if err != nil {
				logger.Log.Error("Error reading message: ", err.Error())
				return err
			}
			logger.Log.Debug("Received message from Kafka topic: ", m.Topic, " Partition: ", m.Partition, " Offset: ", m.Offset, " Key: ", string(m.Key), " Value: ", string(m.Value))

			// Converte il messaggio in un oggetto HeartbeatMsg
			var heartbeatMsg types.HeartbeatMsg
			heartbeatMsg, err = types.CreateHeartbeatMsgFromKafka(m)
			if err != nil {
				logger.Log.Error("Error unmarshalling Heartbeat Message: ", err.Error())
				continue
			}

			// Prova a inviare il messaggio al canale
			// Se il canale è pieno, ritenta per un numero massimo di volte
			// con un ritardo tra i tentativi
			// Se non riesce a inviare il messaggio, lo scarta e logga un errore
			// per evitare di bloccare il lettore Kafka
			sent := false
			for attempt := 0; attempt < environment.KafkaMaxAttempts && !sent; attempt++ {
				select {
				case heartbeatChannel <- heartbeatMsg:
					// Inviato con successo
					logger.Log.Debug("Heartbeat message sent to channel: ", heartbeatMsg)
					sent = true
				default:
					// Canale pieno, logghiamo un avviso e attendiamo l'elaborazione
					logger.Log.Warn("Heartbeat message channel is full. Attempt(s) ", attempt+1, " of ", environment.KafkaMaxAttempts)
					if attempt == environment.KafkaMaxAttempts-1 {
						logger.Log.Error("Max attempts reached, discarding heartbeat message: ", heartbeatMsg)
						break
					}
					// Attende prima di ritentare
					time.Sleep(time.Duration(environment.KafkaAttemptDelay) * time.Millisecond)
				}
			}
		}
	}
}

// CommitHeartbeatBatchMessages esegue il commit degli offset dei messaggi Kafka in un batch di messaggi di heartbeat.
func CommitHeartbeatBatchMessages(messages []kafka.Message) error {
	// Se il lettore Kafka non è inizializzato, non fare nulla
	if kafkaHeartbeatReader == nil {
		return nil
	}

	if len(messages) == 0 {
		return nil
	}

	// Esegue il commit dei messaggi
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(environment.KafkaCommitTimeout)*time.Second)
	defer cancel()

	err := kafkaHeartbeatReader.CommitMessages(ctx, messages...)
	if err != nil {
		logger.Log.Error("Failed to commit Kafka messages: ", err)
		return err
	}

	logger.Log.Debug("Committed Kafka ", len(messages), " messages")
	return nil
}
