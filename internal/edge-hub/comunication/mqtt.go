package comunication

import (
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// sensorClient è il client MQTT utilizzato per ricevere i dati dai sensori.
// Questo client è utilizzato per comunicare con i sensori e ricevere i dati in tempo reale.
var sensorClient MQTT.Client

// hubClient è il client MQTT utilizzato per pubblicare i dati filtrati e le configurazioni.
// Questo client è utilizzato per comunicare con il Proximity Hub.
var hubClient MQTT.Client

// Contatori per i tentativi di connessione
var connectAttempts = 0

// sensorDataHandler è la funzione di callback che processa i messaggi con le rilevazioni in arrivo.
func makeSensorDataHandler(sensorDataChannel chan types.SensorData) MQTT.MessageHandler {
	return func(client MQTT.Client, msg MQTT.Message) {
		logger.Log.Debug("Received message on topic: ", msg.Topic())

		sensorData, err := types.CreateSensorDataFromMQTT(msg)
		if err != nil {
			logger.Log.Error("Error parsing sensor data from MQTT message: ", err.Error())
			return
		}

		// select non bloccante, il ricevitore MQTT non viene mai bloccato dal processore
		// Se il canale è pieno, scarta il messaggio per non rallentare la ricezione.
		select {

		case sensorDataChannel <- sensorData:
			// Messaggio inviato correttamente
			logger.Log.Debug("Sent message on sensorDataChannel")
		default:
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

		if configMsg == (types.ConfigurationMsg{}) {
			logger.Log.Debug("Received nil configuration message. Skipping processing.")
			return
		}

		// select non bloccante, il ricevitore MQTT non viene mai bloccato dal processore
		// Se il canale è pieno, scarta il messaggio per non rallentare la ricezione.
		select {
		case configurationMessageChannel <- configMsg:
			// Messaggio inviato correttamente
		default:
			logger.Log.Warn("Data channel is full. Discarding message from sensor: ", configMsg.SensorID)
		}
	}
}

// makeConnectionHandler viene chiamata quando la connessione MQTT è già riuscita e quello che facciamo ora è
// fare la subscribe al topic dei dati del sensore. Cioè praticamente stiamo dicendo che l edge hub è sottoscritto
// alla ricezione dei dati da parte dei sensori e quindi li riceve
func makeConnectionHandler(sensorDataChannel chan types.SensorData, configurationMessageChannel chan types.ConfigurationMsg) MQTT.OnConnectHandler {
	return func(client MQTT.Client) {

		var topic string
		var token MQTT.Token

		if sensorDataChannel != nil && (environment.ServiceMode == types.EdgeHubService || environment.ServiceMode == types.EdgeHubFilterService) {
			topic = environment.SensorDataTopic + "/#"
			logger.Log.Debug("Subscribing to topic: ", topic)

			// Sottoscrivi al topic per ricevere i dati dai sensori
			// QoS 0, cioè "at most once", il messaggio può andare perso
			token = client.Subscribe(topic, 0, makeSensorDataHandler(sensorDataChannel)) // Il message handler è globale
			logger.Log.Info("Subscribed to topic: ", topic)
			if token.WaitTimeout(time.Duration(environment.MaxSubscriptionTimeout)*time.Second) && token.Error() != nil {
				logger.Log.Error("Failed to subscribe to topic:", topic, "error:", token.Error())
				os.Exit(1) // Esci se non riesci a sottoscrivere
			}
		}

		if configurationMessageChannel != nil && (environment.ServiceMode == types.EdgeHubService || environment.ServiceMode == types.EdgeHubConfigurationService) {
			topic = environment.SensorConfigurationTopic + "/#"
			logger.Log.Debug("Subscribing to topic: ", topic)

			// Sottoscrivi al topic per ricevere le configurazioni dai sensori
			// QoS 2, cioè "exactly once", il messaggio viene consegnato una sola volta, senza duplicati
			token = client.Subscribe(topic, 2, makeConfigurationMessageHandler(configurationMessageChannel)) // Il message handler è globale
			logger.Log.Info("Subscribed to topic: ", topic)
			if token.WaitTimeout(time.Duration(environment.MaxSubscriptionTimeout)*time.Second) && token.Error() != nil {
				logger.Log.Error("Failed to subscribe to topic:", topic, "error:", token.Error())
				os.Exit(1) // Esci se non riesci a sottoscrivere
			}
		}

		// Se siamo qui, la connessione è riuscita e abbiamo sottoscritto ai topic
		// Quindi resettiamo il contatore dei tentativi di connessione
		connectAttempts = 0
		logger.Log.Info("Successfully connected to MQTT broker.")
	}
}

// getCommonOptions restituisce le opzioni comuni per entrambi i client MQTT.
// Queste opzioni includono le impostazioni di connessione, resilienza e callback.
// Queste opzioni sono condivise tra il client dei sensori e il client del Proximity Hub.
// In questo modo, possiamo evitare di duplicare il codice e mantenere una configurazione
// coerente tra i due client.
// Le opzioni comuni includono:
// - ID del client basato su EdgeMacrozone, EdgeZone e HubID
// - Protocollo MQTT versione 5
// - Auto riconnessione abilitata
// - Intervallo massimo di riconnessione
// - Gestione della connessione riuscita con la sottoscrizione ai topic dei dati
// - Gestione della connessione riuscita con la sottoscrizione ai topic di configurazione
func getCommonOptions(sensorDataChannel chan types.SensorData, configurationMessageChannel chan types.ConfigurationMsg) *MQTT.ClientOptions {

	// --- Impostazioni di Connessione ---

	mqttId := environment.EdgeMacrozone + "_" + environment.EdgeZone + "_" + environment.HubID
	logger.Log.Debug(fmt.Sprintf("MQTT Client ID: %s", mqttId))

	opts := MQTT.NewClientOptions()
	opts.SetClientID(mqttId)
	opts.SetProtocolVersion(5)

	// --- Impostazioni di Resilienza ---

	// la libreria paho gestisce automaticamente la riconnessione in background,
	opts.SetAutoReconnect(true)
	// controlla la frequenza di tentativi della riconnessione
	logger.Log.Debug("Setting MQTT Max Reconnect Interval to ", environment.MaxReconnectionInterval, " seconds")
	opts.SetMaxReconnectInterval(time.Duration(environment.MaxReconnectionInterval) * time.Second)
	opts.SetConnectRetry(true)

	// Imposta il callback per la connessione riuscita
	// Questo handler viene chiamato quando la connessione è stabilita con successo
	// e permette di sottoscrivere ai topic desiderati.
	// In questo caso, sottoscrive al topic dei dati e di configurazione del sensore.
	opts.SetOnConnectHandler(makeConnectionHandler(sensorDataChannel, configurationMessageChannel))
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
		if connectAttempts > environment.MaxReconnectionAttempts {
			logger.Log.Error("Max connection attempts reached. Exiting.")
			os.Exit(1)
		}
		logger.Log.Warn("Hub attempting to connect to MQTT broker: ", connectAttempts, " attempt(s) on ", environment.MaxReconnectionAttempts, " max attempt(s)")
		return tlsCfg
	})

	return opts

}

// connectAndManage gestisce la connessione una sola volta.
// Se il client è già definito e connesso, non fa nulla.
// Se il client non è definito o non è connesso, procede con la connessione al broker MQTT.
func connectAndManage(sensorDataChannel chan types.SensorData, configurationMessageChannel chan types.ConfigurationMsg) {

	sensorBrokerURL := fmt.Sprintf("%s://%s:%s", environment.MqttSensorBrokerProtocol, environment.MqttSensorBrokerAddress, environment.MqttSensorBrokerPort)
	hubBrokerURL := fmt.Sprintf("%s://%s:%s", environment.MqttHubBrokerProtocol, environment.MqttHubBrokerAddress, environment.MqttHubBrokerPort)

	if sensorClient == nil || (sensorClient != nil && !sensorClient.IsConnected()) {

		// Se il broker del Proximity Hub è lo stesso del broker dei sensori,
		// riutilizza il client del Proximity Hub per i sensori.
		if sensorBrokerURL == hubBrokerURL && hubClient != nil && hubClient.IsConnected() {
			sensorClient = hubClient
			logger.Log.Info("Reusing hub client for sensor connection.")
			return
		}

		opts := getCommonOptions(sensorDataChannel, configurationMessageChannel)

		// --- Connessione al broker ---

		logger.Log.Debug(fmt.Sprintf("Broker URL: %s", sensorBrokerURL))
		opts.AddBroker(sensorBrokerURL)

		sensorClient = MQTT.NewClient(opts)

		logger.Log.Info("Hub attempting to connect to MQTT sensor broker at ", sensorBrokerURL)
		if token := sensorClient.Connect(); token.WaitTimeout(time.Duration(environment.MaxReconnectionTimeout)*time.Second) && token.Error() != nil {
			// L'errore di connessione inziale viene solo loggato, il meccanismo di
			// AutoReconnect continuerà a tentare in background
			logger.Log.Error("Hub failed to connect initially:", token.Error())
		}

	}

	if hubClient == nil || (hubClient != nil && !hubClient.IsConnected()) {

		// Se il broker del Proximity Hub è lo stesso del broker dei sensori,
		// riutilizza il client dei sensori per il Proximity Hub.
		if sensorBrokerURL == hubBrokerURL && sensorClient != nil && sensorClient.IsConnected() {
			hubClient = sensorClient
			logger.Log.Info("Reusing sensor client for Proximity Hub connection.")
			return
		}

		opts := getCommonOptions(sensorDataChannel, configurationMessageChannel)

		// --- Connessione al broker ---

		logger.Log.Debug(fmt.Sprintf("Broker URL: %s", hubBrokerURL))
		opts.AddBroker(hubBrokerURL)

		hubClient = MQTT.NewClient(opts)

		logger.Log.Info("Hub attempting to connect to MQTT hub broker at ", hubBrokerURL)
		if token := hubClient.Connect(); token.WaitTimeout(time.Duration(environment.MaxReconnectionTimeout)*time.Second) && token.Error() != nil {
			// L'errore di connessione inziale viene solo loggato, il meccanismo di
			// AutoReconnect continuerà a tentare in background
			logger.Log.Error("Hub failed to connect initially:", token.Error())
		}

	}
}

func SetupMQTTConnection(sensorDataChannel chan types.SensorData, configurationMessageChannel chan types.ConfigurationMsg) {

	// Assicura che la connessione non sia già stata inizializzata.
	if sensorClient != nil && sensorClient.IsConnected() && hubClient != nil && hubClient.IsConnected() {
		logger.Log.Info("MQTT clients already connected. Skipping setup.")
		return
	}

	// Inizializza la connessione MQTT
	connectAndManage(sensorDataChannel, configurationMessageChannel)

	// Non procedere se la connessione non è attiva.
	if !sensorClient.IsConnected() {
		logger.Log.Error("MQTT sensor client not connected.")
		os.Exit(1)
	}

	// Non procedere se la connessione non è attiva.
	if !hubClient.IsConnected() {
		logger.Log.Error("MQTT hub client not connected.")
		os.Exit(1)
	}

}

// PublishFilteredData pubblica i dati filtrati al broker MQTT
// Praticamente sopra abbiamo fatto la subscribe nel ricevere i dati dai sensori, ora invece
// facciamo diventare l edge hub un attore che pubblica i dati filtrati (che saranno presi dal
// proximity fog, il quale farà a sua volta la subscribe al broker mqtt per ricevere i dati)
func PublishFilteredData(filteredDataChannel chan types.SensorData) {

	// Non procedere se la connessione non è attiva.
	if !hubClient.IsConnected() {
		logger.Log.Warn("MQTT client not connected. Skipping data publishing.")
		// L'opzione AutoReconnect della libreria sta già lavorando per riconnettersi.
		return
	}

	for filteredData := range filteredDataChannel {
		payload, err := json.Marshal(filteredData)
		if err != nil {
			logger.Log.Error("Error during JSON serialization: ", err.Error())
			return
		}

		topic := environment.FilteredDataTopic + "/" + filteredData.SensorID

		// Invia i dati al broker MQTT
		// QoS 0, cioè "at most once", il messaggio può andare perso
		// Retained false, il messaggio non viene conservato dal broker
		//
		// Usiamo WaitTimeout per non bloccare l'hub all'infinito,
		// cioè se la rete è lenta l'hub comunque non si blocca
		// anche se il timeout scade e si raggiunge il max dei
		// retry il programma comunque prosegue
		for i := 0; i < environment.MessagePublishAttempts; i++ {
			token := hubClient.Publish(topic, 0, false, payload)
			if !token.WaitTimeout(time.Duration(environment.MessagePublishTimeout) * time.Second) {
				logger.Log.Warn("Timeout publishing message. Retry ", i+1)
			} else if err := token.Error(); err != nil {
				logger.Log.Error("Error publishing message: ", err.Error(), ". Retry ", i+1)
			} else {
				logger.Log.Debug("Message published successfully on topic: ", topic)
				break
			}
		}
	}
}

func PublishConfigurationMessage(configurationMessageChannel chan types.ConfigurationMsg) {

	// Non procedere se la connessione non è attiva.
	if !hubClient.IsConnected() {
		logger.Log.Warn("MQTT client not connected. Skipping data publishing.")
		// L'opzione AutoReconnect della libreria sta già lavorando per riconnettersi.
		return
	}

	for msg := range configurationMessageChannel {
		payload, err := json.Marshal(msg)
		if err != nil {
			logger.Log.Error("Error during JSON serialization: ", err.Error())
			return
		}

		topic := environment.HubConfigurationTopic + "/" + msg.SensorID

		// Invia i dati al broker MQTT
		// QoS 2, cioè "exactly once", il messaggio viene consegnato una sola volta, senza duplicati
		// Retained true, il broker conserva l’ultimo messaggio pubblicato su un topic e lo invia ai nuovi iscritti
		//
		// Usiamo WaitTimeout per non bloccare l'hub all'infinito,
		// cioè se la rete è lenta l'hub comunque non si blocca
		// anche se il timeout scade e si raggiunge il max dei
		// retry il programma comunque prosegue
		for i := 0; i < environment.MessagePublishAttempts; i++ {
			token := hubClient.Publish(topic, 2, true, payload)
			if !token.WaitTimeout(time.Duration(environment.MessagePublishTimeout) * time.Second) {
				logger.Log.Warn("Timeout publishing message. Retry ", i+1)
			} else if err := token.Error(); err != nil {
				logger.Log.Error("Error publishing message: ", err.Error(), ". Retry ", i+1)
			} else {
				logger.Log.Debug("Message published successfully on topic: ", topic)
				break
			}
		}
	}
}

// CleanRetentionConfigurationMessage Rimuove il messaggio di configurazione dal canale se è già stato elaborato.
// Questo è utile per evitare di elaborare più volte lo stesso messaggio.
func CleanRetentionConfigurationMessage(msg types.ConfigurationMsg) {

	// Pulisce il messaggio di configurazione solo se è un messaggio di nuovo sensore
	// e quindi non è più necessario mantenere il messaggio di configurazione.
	if msg.MsgType == types.NewSensorMsgType {
		logger.Log.Debug("Cleaning retention for configuration message: ", msg)

		// Non procedere se la connessione non è attiva.
		if !sensorClient.IsConnected() {
			logger.Log.Warn("MQTT client not connected. Skipping data cleaning.")
			// L'opzione AutoReconnect della libreria sta già lavorando per riconnettersi.
			return
		}

		// Per pulire un messaggio di configurazione, pubblichiamo un messaggio vuoto
		// sullo stesso topic con retained=true. Questo indica al broker di rimuovere
		// il messaggio precedente.
		topic := environment.SensorConfigurationTopic + "/" + msg.SensorID
		token := sensorClient.Publish(topic, 2, true, "")

		if !token.WaitTimeout(time.Duration(environment.MessageCleaningTimeout) * time.Second) {
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
}

// SendRegistrationMessage invia un messaggio di registrazione al broker MQTT
// per registrare l'hub con le sue informazioni di configurazione.
func SendRegistrationMessage() {

	// Non procedere se il messaggio di configurazione non viene inviato
	for {

		if !hubClient.IsConnected() {
			logger.Log.Warn("MQTT client not connected. Skipping data publishing.")
			// L'opzione AutoReconnect della libreria sta già lavorando per riconnettersi.
			// Aspetta prima di riprovare
			time.Sleep(time.Duration(environment.MaxReconnectionInterval) * time.Second)
			continue
		}

		// Crea il messaggio di configurazione
		payload, err := json.Marshal(types.ConfigurationMsg{
			EdgeMacrozone: environment.EdgeMacrozone,
			MsgType:       types.NewEdgeMsgType,
			Timestamp:     time.Now().UTC().Unix(),
			Service:       environment.ServiceMode,
			EdgeZone:      environment.EdgeZone,
			HubID:         environment.HubID,
		})
		if err != nil {
			logger.Log.Error("Error during JSON serialization: ", err.Error())
			os.Exit(1)
		}

		// Invia i dati al broker MQTT
		// QoS 2, cioè "exactly once", il messaggio viene consegnato una sola volta, senza duplicati
		// Retained true, il broker conserva l’ultimo messaggio pubblicato su un topic e lo invia ai nuovi iscritti
		topic := environment.HubConfigurationTopic + "/" + environment.HubID
		token := hubClient.Publish(topic, 2, true, payload)

		// Usiamo Wait per aspettare il completamento della pubblicazione
		// Non usiamo WaitTimeout perché vogliamo essere sicuri che il messaggio
		// venga inviato prima di procedere.
		// Se il messaggio non viene inviato, riproviamo.
		if !token.Wait() {
			logger.Log.Warn("Timeout publishing configuration message.")
			// Aspetta prima di riprovare
			time.Sleep(time.Duration(environment.MaxReconnectionInterval) * time.Second)
			continue
		} else if err := token.Error(); err != nil {
			logger.Log.Error("Error publishing configuration message: ", err.Error())
			os.Exit(1)
		} else {
			logger.Log.Debug("Configuration message published successfully on topic: ", topic)
			return
		}

	}
}

// SendHeartbeatMessage invia un messaggio di heartbeat al broker MQTT
// per mantenere viva la connessione e segnalare che l'hub è attivo
func SendHeartbeatMessage() {

	msg := types.HeartbeatMsg{
		EdgeMacrozone: environment.EdgeMacrozone,
		EdgeZone:      environment.EdgeZone,
		HubID:         environment.HubID,
	}

	topic := environment.HeartbeatTopic + "/" + environment.HubID

	// L'heartbeat viene inviato periodicamente per mantenere viva la connessione
	for {

		// Non procedere se la connessione non è attiva.
		if !hubClient.IsConnected() {
			logger.Log.Warn("MQTT client not connected. Skipping heartbeat message publishing.")
			// L'opzione AutoReconnect della libreria sta già lavorando per riconnettersi.
			return
		}

		// Crea il messaggio di heartbeat
		msg.Timestamp = time.Now().UTC().Unix()
		payload, err := json.Marshal(msg)
		if err != nil {
			logger.Log.Error("Error during JSON serialization: ", err.Error())
			os.Exit(1)
		}

		// Invia i dati al broker MQTT
		// QoS 1, cioè "at least once", il messaggio viene consegnato almeno una volta, può essere duplicato
		// Retained true, il broker conserva l’ultimo messaggio pubblicato su un topic e lo invia ai nuovi iscritti
		token := hubClient.Publish(topic, 1, true, payload)

		// Usiamo WaitTimeout per non attendere all'infinito
		if !token.Wait() {
			logger.Log.Warn("Timeout publishing heartbeat message.")
			// Aspetta prima di riprovare
			time.Sleep(time.Duration(environment.MaxReconnectionInterval) * time.Second)
			continue
		} else if err := token.Error(); err != nil {
			logger.Log.Error("Error publishing heartbeat message: ", err.Error())
			os.Exit(1)
		} else {
			logger.Log.Debug("Heartbeat message published successfully on topic: ", topic)
			time.Sleep(environment.HeartbeatInterval)
		}

	}

}
