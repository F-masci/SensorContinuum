package comunication

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// client è il client MQTT globale utilizzato per la connessione al broker.
var client MQTT.Client

// connectAttempts Contatori per i tentativi di connessione
var connectAttempts = 0

// sensorDataHandler è la funzione di callback che processa i messaggi in arrivo.
func makeSensorDataHandler(filteredDataChannel chan types.SensorData) MQTT.MessageHandler {
	return func(client MQTT.Client, msg MQTT.Message) {
		logger.Log.Debug("Received message on topic: ", msg.Topic())

		// convertiamo il messaggio grezzo MQTT nella struttura dati SensorData
		sensorData, err := types.CreateSensorDataFromMQTT(msg)
		if err != nil {
			logger.Log.Error("Error parsing sensor data from MQTT message: ", err.Error())
			return
		}

		// select non bloccante, il ricevitore MQTT non viene mai bloccato dal processore
		// Se il canale è pieno, scarta il messaggio per non rallentare la ricezione.
		select {
		case filteredDataChannel <- sensorData:
			// Messaggio inviato correttamente
			logger.Log.Debug("Sent message to filteredDataChannel")
		default:
			// Se il canale è pieno scartiamo il messaggio per non bloccare la ricezione dei nuovi
			logger.Log.Warn("Data channel is full. Discarding message from sensor: ", sensorData.SensorID)
		}
	}
}

// configurationMessageHandler è la funzione di callback che processa i messaggi di configurazione in arrivo.
func makeConfigurationMessageHandler(configurationMessageChannel chan types.ConfigurationMsg) MQTT.MessageHandler {
	return func(client MQTT.Client, msg MQTT.Message) {
		logger.Log.Debug("Received message on topic: ", msg.Topic())

		// convertiamo il messaggio grezzo MQTT nella struttura dati ConfigurationMsg
		configMsg, err := types.CreateConfigurationMsgFromMqtt(msg)
		if err != nil {
			logger.Log.Error("Error parsing sensor data from MQTT message: ", err.Error())
			return
		}

		// select non bloccante, il ricevitore MQTT non viene mai bloccato dal processore
		// Se il canale è pieno, scarta il messaggio per non rallentare la ricezione.
		select {
		case configurationMessageChannel <- configMsg:
			// Messaggio inviato correttamente
			logger.Log.Debug("Sent configuration message to channel")
		default:
			logger.Log.Warn("Configuration message channel is full. Discarding message: ", configMsg)
		}
	}
}

// heartbeatMessageHandler è la funzione di callback che processa i messaggi di heartbeat in arrivo.
func makeHeartbeatMessageHandler(heartbeatMessageChannel chan types.HeartbeatMsg) MQTT.MessageHandler {
	return func(client MQTT.Client, msg MQTT.Message) {
		logger.Log.Debug("Received message on topic: ", msg.Topic())

		// convertiamo il messaggio grezzo MQTT nella struttura dati HeartbeatMsg
		heartbeatMsg, err := types.CreateHeartbeatMsgFromMqtt(msg)
		if err != nil {
			logger.Log.Error("Error parsing sensor data from MQTT message: ", err.Error())
			return
		}

		// select non bloccante, il ricevitore MQTT non viene mai bloccato dal processore
		// Se il canale è pieno, scarta il messaggio per non rallentare la ricezione.
		select {
		case heartbeatMessageChannel <- heartbeatMsg:
			// Messaggio inviato correttamente
			logger.Log.Debug("Sent heartbeat message to channel")
		default:
			logger.Log.Warn("Heartbeat message channel is full. Discarding message: ", heartbeatMsg)
		}
	}
}

// makeConnectionHandler viene chiamata quando la connessione MQTT è già riuscita e quello che fa ora è
// sottoscriversi ai topic desiderati.
func makeConnectionHandler(filteredDataChannel chan types.SensorData, configurationMessageChannel chan types.ConfigurationMsg, heartbeatMessageChannel chan types.HeartbeatMsg) MQTT.OnConnectHandler {
	return func(client MQTT.Client) {

		var topic string
		var token MQTT.Token

		if filteredDataChannel != nil && (environment.ServiceMode == types.ProximityHubLocalCacheService || environment.ServiceMode == types.ProximityHubService) {
			topic = environment.FilteredDataTopic + "/#"
			logger.Log.Debug("Subscribing to topic: ", topic)

			// Sottoscrivi al topic per ricevere i dati dai sensori
			// QoS 0, cioè "at most once", il messaggio può andare perso
			token = client.Subscribe(topic, 0, makeSensorDataHandler(filteredDataChannel)) // Il message handler è globale
			logger.Log.Info("Subscribed to topic: ", topic)
			if token.WaitTimeout(time.Duration(environment.MqttMaxSubscriptionTimeout)*time.Second) && token.Error() != nil {
				logger.Log.Error("Failed to subscribe to topic:", topic, "error:", token.Error())
				os.Exit(1) // Esci se non riesci a sottoscrivere
			}
		}

		if configurationMessageChannel != nil && (environment.ServiceMode == types.ProximityHubConfigurationService || environment.ServiceMode == types.ProximityHubService) {
			topic = environment.HubConfigurationTopic + "/#"
			logger.Log.Debug("Subscribing to topic: ", topic)

			// Sottoscrivi al topic per ricevere le configurazioni dai sensori
			// QoS 2, cioè "exactly once", il messaggio viene consegnato una sola volta, senza duplicati
			token = client.Subscribe(topic, 2, makeConfigurationMessageHandler(configurationMessageChannel)) // Il message handler è globale
			logger.Log.Info("Subscribed to topic: ", topic)
			if token.WaitTimeout(time.Duration(environment.MqttMaxSubscriptionTimeout)*time.Second) && token.Error() != nil {
				logger.Log.Error("Failed to subscribe to topic:", topic, "error:", token.Error())
				os.Exit(1) // Esci se non riesci a sottoscrivere
			}
		}

		if heartbeatMessageChannel != nil && (environment.ServiceMode == types.ProximityHubHeartbeatService || environment.ServiceMode == types.ProximityHubService) {
			topic = environment.HeartbeatTopic + "/#"
			logger.Log.Debug("Subscribing to topic: ", topic)

			// Sottoscrivi al topic per ricevere i messaggi di heartbeat
			// QoS 1, cioè "at least once", il messaggio viene consegnato almeno una volta, possono esserci duplicati
			token = client.Subscribe(topic, 1, makeHeartbeatMessageHandler(heartbeatMessageChannel)) // Il message handler è globale
			logger.Log.Info("Subscribed to topic: ", topic)
			if token.WaitTimeout(time.Duration(environment.MqttMaxSubscriptionTimeout)*time.Second) && token.Error() != nil {
				logger.Log.Error("Failed to subscribe to topic:", topic, "error:", token.Error())
				os.Exit(1) // Esci se non riesci a sottoscrivere
			}

		}

		// Se siamo qui, la connessione è riuscita e abbiamo sottoscritto ai topic
		// Quindi resettiamo il contatore dei tentativi di connessione
		connectAttempts = 0
		logger.Log.Info("Successfully subscribed and connected to MQTT broker.")

	}
}

// connectAndManage gestisce la connessione al broker MQTT e la riconnessione in caso di perdita della connessione.
// Se la connessione è già attiva, non fa nulla.
func connectAndManage(filteredDataChannel chan types.SensorData, configurationMessageChannel chan types.ConfigurationMsg, heartbeatMessageChannel chan types.HeartbeatMsg) {
	if client != nil && client.IsConnected() {
		return
	}

	mqttId := environment.EdgeMacrozone + "_" + environment.HubID
	brokerURL := fmt.Sprintf("%s://%s:%s", environment.MqttProtocol, environment.MqttBroker, environment.MqttPort)

	logger.Log.Info("MQTT configuration: clientID ", mqttId, " - brokerURL ", brokerURL)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(mqttId)
	opts.SetProtocolVersion(5)

	// --- Impostazioni di Resilienza ---

	// la libreria paho gestisce automaticamente la riconnessione in background,
	opts.SetAutoReconnect(true)
	// controlla la frequenza di tentativi della riconnessione
	logger.Log.Debug("Setting MQTT Max Reconnect Interval to ", environment.MqttMaxReconnectionInterval, " seconds")
	opts.SetMaxReconnectInterval(time.Duration(environment.MqttMaxReconnectionInterval) * time.Second)
	opts.SetConnectRetry(true)

	// Imposta il callback per la connessione riuscita
	// Questo handler viene chiamato quando la connessione è stabilita con successo
	// e permette di sottoscrivere ai topic desiderati.
	// In questo caso, sottoscrive al topic dei dati, di configurazione del sensore
	// e di heartbeat.
	opts.SetOnConnectHandler(makeConnectionHandler(filteredDataChannel, configurationMessageChannel, heartbeatMessageChannel))
	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		logger.Log.Warn("Hub lost connection to MQTT broker: ", err.Error())
	})

	// Limita il numero di tentativi di connessione
	// Se il numero di tentativi supera maxConnectAttempts, il programma termina.
	//
	// Questo è utile per evitare loop infiniti in caso di problemi di connessione persistenti
	// e per evitare che l'hub continui a tentare di connettersi
	// in un ciclo infinito senza successo.
	opts.SetConnectionAttemptHandler(func(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
		connectAttempts++
		if connectAttempts > environment.MqttMaxReconnectionAttempts {
			logger.Log.Error("Max connection attempts reached. Exiting.")
			os.Exit(1)
		}
		logger.Log.Warn("Hub attempting to connect to MQTT broker: ", connectAttempts, " attempt(s) on ", environment.MqttMaxReconnectionAttempts, " max attempt(s)")
		return tlsCfg
	})

	client = MQTT.NewClient(opts)

	logger.Log.Info("Hub attempting to connect to MQTT sensor broker at ", brokerURL)
	if token := client.Connect(); token.WaitTimeout(time.Duration(environment.MqttMaxReconnectionTimeout)*time.Second) && token.Error() != nil {
		// L'errore di connessione iniziale viene solo loggato, il meccanismo di
		// AutoReconnect continuerà a tentare in background
		logger.Log.Error("Hub failed to connect initially:", token.Error())
	}
}

func SetupMQTTConnection(filteredDataChannel chan types.SensorData, configurationMessageChannel chan types.ConfigurationMsg, heartbeatMessageChannel chan types.HeartbeatMsg) {

	// Assicura che la connessione non sia già stata inizializzata.
	if client != nil && client.IsConnected() {
		logger.Log.Info("MQTT client already connected. Skipping setup.")
		return
	}

	// Inizializza la connessione MQTT
	connectAndManage(filteredDataChannel, configurationMessageChannel, heartbeatMessageChannel)

	// Non procedere se la connessione non è attiva.
	if !client.IsConnected() {
		logger.Log.Error("MQTT client not connected.")
		os.Exit(1)
	}
}

// CleanRetentionConfigurationMessage Rimuove il messaggio di configurazione dal canale se è già stato elaborato.
// Questo è utile per evitare di elaborare più volte lo stesso messaggio.
func CleanRetentionConfigurationMessage(msg types.ConfigurationMsg) {

	logger.Log.Debug("Cleaning retention for configuration message: ", msg)

	// Non procedere se la connessione non è attiva.
	if !client.IsConnected() {
		logger.Log.Warn("MQTT not connected. Skipping data cleaning.")
		// L'opzione AutoReconnect della libreria sta già lavorando per riconnettersi.
		return
	}

	// Per pulire un messaggio di configurazione, pubblichiamo un messaggio vuoto
	// sullo stesso topic con retained=true. Questo indica al broker di rimuovere
	// il messaggio precedente.
	var topic string
	if msg.MsgType == types.NewSensorMsgType {
		topic = environment.HubConfigurationTopic + "/" + msg.EdgeZone + "/" + msg.SensorID
	} else if msg.MsgType == types.NewEdgeMsgType {
		topic = environment.HubConfigurationTopic + "/" + msg.EdgeZone + "/" + msg.HubID
	} else {
		logger.Log.Error("Unknown configuration message type, cannot clean retention.")
		return
	}
	// invia i dati al broker MQTT
	token := client.Publish(topic, 2, true, "")

	if !token.WaitTimeout(time.Duration(environment.MqttMessageCleaningTimeout) * time.Second) {
		logger.Log.Error("Timeout cleaning message: ", topic)
		return
	}
	err := token.Error()
	if err != nil {
		logger.Log.Error("Error cleaning message: ", err.Error())
		return
	}
	logger.Log.Debug("Message cleaned successfully: ", topic)
}

// CleanRetentionHeartbeatMessage Rimuove il messaggio di heartbeat dal canale se è già stato elaborato.
// Questo è utile per evitare di elaborare più volte lo stesso messaggio.
func CleanRetentionHeartbeatMessage(msg types.HeartbeatMsg) {

	logger.Log.Debug("Cleaning retention for heartbeat message: ", msg)

	// Non procedere se la connessione non è attiva.
	if !client.IsConnected() {
		logger.Log.Warn("MQTT client not connected. Skipping data cleaning.")
		// L'opzione AutoReconnect della libreria sta già lavorando per riconnettersi.
		return
	}

	// Per pulire un messaggio di heartbeat, pubblichiamo un messaggio vuoto
	// sullo stesso topic con retained=true. Questo indica al broker di rimuovere
	// il messaggio precedente.
	topic := environment.HeartbeatTopic + "/" + msg.EdgeZone + "/" + msg.HubID
	token := client.Publish(topic, 2, true, "")

	if !token.WaitTimeout(time.Duration(environment.MqttMessageCleaningTimeout) * time.Second) {
		logger.Log.Error("Timeout cleaning message: ", topic)
		return
	}
	err := token.Error()
	if err != nil {
		logger.Log.Error("Error cleaning message: ", err.Error())
		return
	}
	logger.Log.Debug("Message cleaned successfully: ", topic)
}
