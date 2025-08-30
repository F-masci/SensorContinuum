package comunication

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

// realtimeKafkaWriter per le misurazioni in tempo reale
var realtimeKafkaWriter *kafka.Writer = nil

// statsKafkaWriter per le statistiche aggregate
var statsKafkaWriter *kafka.Writer = nil

// configurationKafkaWriter per i messaggi di configurazione
var configurationKafkaWriter *kafka.Writer = nil

// heartbeatKafkaWriter per i messaggi di heartbeat
var heartbeatKafkaWriter *kafka.Writer = nil

// connect stabilisce la connessione con Kafka se non è già stabilita
func connect() {

	// Se tutte le connessioni sono già stabilite, non fare nulla
	if realtimeKafkaWriter != nil && statsKafkaWriter != nil && configurationKafkaWriter != nil {
		return
	}

	// ------ KAFKA ACK ------
	// RequireOne: il leader del topic deve confermare la ricezione del messaggio prima di considerarlo inviato con successo.
	// RequireAll: tutti i repliche del topic devono confermare la ricezione del messaggio prima di considerarlo inviato con successo.
	// RequireNone: non è necessaria alcuna conferma di ricezione, il messaggio è considerato inviato con successo non appena viene scritto nel buffer del client Kafka.
	// ----------------------

	// Connessione per il topic delle misurazioni in tempo reale
	realtimeKafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityRealtimeDataTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (write) to Kafka topic for real-time data, topic: ", environment.ProximityRealtimeDataTopic)

	// Connessione per il topic delle statistiche
	statsKafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityAggregatedStatsTopic,
		RequiredAcks: kafka.RequireAll,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (write) to Kafka topic for stats data, topic: ", environment.ProximityAggregatedStatsTopic)

	// Connessione per il topic della configurazione
	configurationKafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityConfigurationTopic,
		RequiredAcks: kafka.RequireAll,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (write) to Kafka topic for configuration data, topic: ", environment.ProximityConfigurationTopic)

	// Connessione per il topic dei messaggi di heartbeat
	heartbeatKafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(environment.KafkaBroker + ":" + environment.KafkaPort),
		Topic:        environment.ProximityHeartbeatTopic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.Hash{},
	}
	logger.Log.Info("Connected (write) to Kafka topic for heartbeat messages, topic: ", environment.ProximityHeartbeatTopic)
}

// SendRealTimeData invia i dati del sensore al topic Kafka dedicato
func SendRealTimeData(dataBatch []types.SensorData) error {
	// Assicuriamoci di essere connessi a Kafka
	connect()

	// Prepara i messaggi da inviare
	messages := make([]kafka.Message, len(dataBatch))
	for i, data := range dataBatch {
		msgBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		messages[i] = kafka.Message{
			Key:   []byte(environment.EdgeMacrozone),
			Value: msgBytes,
		}
	}

	// Imposta un contesto con timeout per evitare blocchi indefiniti
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(environment.KafkaPublishTimeout)*time.Second)
	defer cancel()

	// Invia i messaggi a Kafka
	return realtimeKafkaWriter.WriteMessages(ctx, messages...)
}

// SendAggregatedData invia le statistiche aggregate al topic Kafka dedicato
func SendAggregatedData(statsBatch []types.AggregatedStats) error {
	// Assicuriamoci di essere connessi a Kafka
	connect()

	// Prepara i messaggi da inviare
	messages := make([]kafka.Message, len(statsBatch))
	for i, data := range statsBatch {
		msgBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		messages[i] = kafka.Message{
			Key:   []byte(environment.EdgeMacrozone),
			Value: msgBytes,
		}
	}

	// Imposta un contesto con timeout per evitare blocchi indefiniti
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(environment.KafkaPublishTimeout)*time.Second)
	defer cancel()

	// Invia i messaggi a Kafka
	return statsKafkaWriter.WriteMessages(ctx, messages...)
}

func SendConfigurationMessage(msg types.ConfigurationMsg) error {
	connect()

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(environment.KafkaPublishTimeout)*time.Second)
	defer cancel()

	return configurationKafkaWriter.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(environment.EdgeMacrozone),
			Value: msgBytes,
		},
	)
}

// SendOwnRegistrationMessage invia il messaggio di registrazione del Proximity Fog Hub al Intermediate Fog Hub
func SendOwnRegistrationMessage() {

	// Non procedere se il messaggio di configurazione non viene inviato
	for {

		// Crea il messaggio di registrazione
		msg := types.ConfigurationMsg{
			MsgType:       types.NewProximityMsgType,
			EdgeMacrozone: environment.EdgeMacrozone,
			Timestamp:     time.Now().UTC().Unix(),
			HubID:         environment.HubID,
			Service:       types.ProximityHubService,
		}

		// Invia il messaggio di registrazione
		if err := SendConfigurationMessage(msg); err != nil {
			logger.Log.Error("Failed to send own registration message: ", err)
			os.Exit(1)
		}

		logger.Log.Info("Own registration message sent successfully.")
		return

	}
}

// SendHeartbeatMessage invia un messaggio di heartbeat al topic Kafka dedicato
func SendHeartbeatMessage(msg types.HeartbeatMsg) error {
	connect()

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(environment.KafkaPublishTimeout)*time.Second)
	defer cancel()

	key := msg.EdgeMacrozone + "-" + msg.EdgeZone + "-" + msg.HubID
	return heartbeatKafkaWriter.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(key),
			Value: msgBytes,
		},
	)

}

// SendOwnHeartbeatMessage invia periodicamente un messaggio di heartbeat al Intermediate Fog Hub
func SendOwnHeartbeatMessage() {
	for {
		logger.Log.Info("Sending own heartbeat message to Intermediate Fog Hub...")

		heartbeatMsg := types.HeartbeatMsg{
			EdgeMacrozone: environment.EdgeMacrozone,
			HubID:         environment.HubID,
			Timestamp:     time.Now().UTC().Unix(),
		}

		if err := SendHeartbeatMessage(heartbeatMsg); err != nil {
			logger.Log.Error("Failed to send own heartbeat message, error: ", err)
		}

		logger.Log.Info("Own heartbeat message sent successfully.")

		time.Sleep(environment.HeartbeatInterval)

	}
}
