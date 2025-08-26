package comunication

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"fmt"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var client MQTT.Client

// sensorDataHandler è la funzione di callback che processa i messaggi in arrivo.
func makeSensorDataHandler(filteredDataChannel chan types.SensorData) MQTT.MessageHandler {
	return func(client MQTT.Client, msg MQTT.Message) {
		logger.Log.Debug("Received message on topic: ", msg.Topic())

		//convertiamo il messaggio grezzo MQTT nella struttura dati SensorData
		sensorData, err := types.CreateSensorDataFromMQTT(msg)
		if err != nil {
			logger.Log.Error("Error parsing sensor data from MQTT message: ", err.Error())
			return
		}

		// tento nel mettere il dato nel canale, se il canale è pieno lo scarto
		select {
		// questa è l'operazione dove metto il dato nel canale "filteredDataChannel"
		// se va a buon fine l'inserimento allora si attiva la goroutine ProcessEdgeHubData avviata dal main
		// che era in attesa di ricevere dati nel canale "filteredDataChannel"
		case filteredDataChannel <- sensorData:
			// Messaggio inviato correttamente, la funzione termina il suo lavoro
			logger.Log.Debug("Sent sensor data to channel")
			break
		default:
			// se il canale è pieno scartiamo il messaggio per non bloccare la ricezione dei nuovi
			logger.Log.Warn("Data channel is full. Discarding message from sensor: ", sensorData.SensorID)
		}
	}
}

// configurationMessageHandler è la funzione di callback che processa i messaggi di configurazione in arrivo.
func makeConfigurationMessageHandler(configurationMessageChannel chan types.ConfigurationMsg) MQTT.MessageHandler {
	return func(client MQTT.Client, msg MQTT.Message) {
		logger.Log.Debug("Received message on topic: ", msg.Topic())

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
			break
		default:
			logger.Log.Warn("Configuration message channel is full. Discarding message: ", configMsg)
		}
	}
}

// heartbeatMessageHandler è la funzione di callback che processa i messaggi di heartbeat in arrivo.
func makeHeartbeatMessageHandler(heartbeatMessageChannel chan types.HeartbeatMsg) MQTT.MessageHandler {
	return func(client MQTT.Client, msg MQTT.Message) {
		logger.Log.Debug("Received message on topic: ", msg.Topic())

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
			break
		default:
			logger.Log.Warn("Heartbeat message channel is full. Discarding message: ", heartbeatMsg)
		}
	}
}

func makeConnectionHandler(filteredDataChannel chan types.SensorData, configurationMessageChannel chan types.ConfigurationMsg, heartbeatMessageChannel chan types.HeartbeatMsg) MQTT.OnConnectHandler {
	return func(client MQTT.Client) {

		topic := environment.FilteredDataTopic + "/#"

		// QoS 0, sottoscrivi al topic per ricevere i dati dai sensori
		token := client.Subscribe(topic, 0, makeSensorDataHandler(filteredDataChannel)) // Il message handler è globale
		logger.Log.Info("Subscribed to topic: ", topic)
		if token.WaitTimeout(5*time.Second) && token.Error() != nil {
			logger.Log.Error("Failed to subscribe to topic:", topic, "error:", token.Error())
			os.Exit(1) // Esci se non riesci a sottoscrivere
		}

		topic = environment.HubConfigurationTopic + "/#"

		// QoS 2, sottoscrivi al topic per ricevere i dati dai sensori
		token = client.Subscribe(topic, 2, makeConfigurationMessageHandler(configurationMessageChannel)) // Il message handler è globale
		logger.Log.Info("Subscribed to topic: ", topic)
		if token.WaitTimeout(5*time.Second) && token.Error() != nil {
			logger.Log.Error("Failed to subscribe to topic:", topic, "error:", token.Error())
			os.Exit(1) // Esci se non riesci a sottoscrivere
		}

		topic = environment.HeartbeatTopic + "/#"

		// QoS 1, sottoscrivi al topic per ricevere i dati dai sensori
		token = client.Subscribe(topic, 1, makeHeartbeatMessageHandler(heartbeatMessageChannel)) // Il message handler è globale
		logger.Log.Info("Subscribed to topic: ", topic)
		if token.WaitTimeout(5*time.Second) && token.Error() != nil {
			logger.Log.Error("Failed to subscribe to topic:", topic, "error:", token.Error())
			os.Exit(1) // Esci se non riesci a sottoscrivere
		}

		logger.Log.Info("Successfully subscribed and connected to MQTT broker")

	}
}

func connectAndManage(filteredDataChannel chan types.SensorData, configurationMessageChannel chan types.ConfigurationMsg, heartbeatMessageChannel chan types.HeartbeatMsg) {
	if client != nil && client.IsConnected() {
		return
	}

	mqttId := environment.EdgeMacrozone + "_" + environment.HubID
	brokerURL := fmt.Sprintf("%s://%s:%s", environment.MosquittoProtocol, environment.MosquittoBroker, environment.MosquittoPort)

	logger.Log.Info("MQTT configuration: clientID ", mqttId, " - brokerURL ", brokerURL)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(mqttId)
	opts.SetProtocolVersion(5)

	// --- Impostazioni di Resilienza ---

	// la libreria paho gestisce automaticamente la riconnessione in background,
	opts.SetAutoReconnect(true)
	//controlla la frequenza dei tentativi di riconnessione
	opts.SetMaxReconnectInterval(10 * time.Second)
	opts.SetConnectRetry(true)

	// Imposta il callback per la connessione riuscita
	// Questo handler viene chiamato quando la connessione è stabilita con successo
	// e permette di sottoscrivere ai topic desiderati.
	// In questo caso, sottoscrive al topic dei dati e di configurazione del sensore.
	opts.SetOnConnectHandler(makeConnectionHandler(filteredDataChannel, configurationMessageChannel, heartbeatMessageChannel))
	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		logger.Log.Warn("Connection lost to MQTT broker: ", err.Error())
	})

	client = MQTT.NewClient(opts)

	logger.Log.Info("Hub attempting to connect to MQTT sensor broker at ", brokerURL)
	if token := client.Connect(); token.WaitTimeout(10*time.Second) && token.Error() != nil {
		//L'errore di connessione inziale viene solo loggato, il meccanismo di
		//AutoReconnect continuerà a tentare in background
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
}

// CleanRetentionConfigurationMessage Rimuove il messaggio di configurazione dal canale se è già stato elaborato.
// Questo è utile per evitare di elaborare più volte lo stesso messaggio.
func CleanRetentionConfigurationMessage(msg types.ConfigurationMsg) {

	if msg.MsgType == types.NewSensorMsgType {
		logger.Log.Debug("Cleaning retention for configuration message: ", msg)

		// Non procedere se la connessione non è attiva.
		if !client.IsConnected() {
			logger.Log.Warn("MQTT client not connected. Skipping data cleaning.")
			// L'opzione AutoReconnect della libreria sta già lavorando per riconnettersi.
			return
		}

		topic := environment.HubConfigurationTopic + "/#"
		// invia i dati al broker MQTT
		token := client.Publish(topic, 2, true, "")

		if !token.Wait() {
			logger.Log.Error("Timeout publishing message.")
			return
		}
		err := token.Error()
		if err != nil {
			logger.Log.Error("Error publishing message: ", err.Error())
			return
		}
		logger.Log.Debug("Message cleaned successfully.")
	}
}
