package comunication

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"fmt"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var client MQTT.Client

// sensorDataHandler è la funzione di callback che processa i messaggi in arrivo.
func makeSensorDataHandler(filteredDataChannel chan structure.SensorData) MQTT.MessageHandler {
	return func(client MQTT.Client, msg MQTT.Message) {
		logger.Log.Debug("Received message on topic: ", msg.Topic())

		sensorData, err := structure.CreateSensorDataFromMQTT(msg)
		if err != nil {
			logger.Log.Error("Error parsing sensor data from MQTT message: ", err.Error())
			return
		}

		// select non bloccante, il ricevitore MQTT non viene mai bloccato dal processore
		// Se il canale è pieno, scarta il messaggio per non rallentare la ricezione.
		select {
		case filteredDataChannel <- sensorData:
			// Messaggio inviato correttamente
			logger.Log.Debug("Sent sensor data to channel")
			break
		default:
			logger.Log.Warn("Data channel is full. Discarding message from sensor: ", sensorData.SensorID)
		}
	}
}

func makeConnectionHandler(filteredDataChannel chan structure.SensorData) MQTT.OnConnectHandler {
	return func(client MQTT.Client) {
		topic := environment.FilteredDataTopic + "/#"
		logger.Log.Info("Successfully connected to MQTT broker. Subscribing to topic: ", topic)
		// QoS 0, sottoscrivi al topic per ricevere i dati
		token := client.Subscribe(topic, 0, makeSensorDataHandler(filteredDataChannel)) // Il message handler è globale
		logger.Log.Info("Subscribed to topic: ", topic)
		if token.WaitTimeout(5*time.Second) && token.Error() != nil {
			logger.Log.Error("Failed to subscribe to topic:", topic, "error:", token.Error())
			os.Exit(1) // Esci se non riesci a sottoscrivere
		}
	}
}

// connectAndManage gestisce la connessione una sola volta.
func connectAndManage(filteredDataChannel chan structure.SensorData) {
	// Se il client è già definito e connesso, non fare nulla.
	if client != nil && client.IsConnected() {
		return
	}

	mqttId := environment.BuildingID + "_" + environment.HubID
	logger.Log.Info(fmt.Sprintf("MQTT Client ID: %s", mqttId))
	brokerURL := fmt.Sprintf("%s://%s:%s", environment.MosquittoProtocol, environment.MosquittoBroker, environment.MosquittoPort)
	logger.Log.Info(fmt.Sprintf("Broker URL: %s", brokerURL))

	opts := MQTT.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(mqttId)
	opts.SetProtocolVersion(5)

	// --- Impostazioni di Resilienza ---

	// la libreria paho gestisce automaticamente la riconnessione in background,
	opts.SetAutoReconnect(true)
	//controlla la frequenza di tenta della riconnessione
	opts.SetMaxReconnectInterval(10 * time.Second)
	opts.SetConnectRetry(true)

	// Imposta il callback per la connessione riuscita
	// Questo handler viene chiamato quando la connessione è stabilita con successo
	// e permette di sottoscrivere ai topic desiderati.
	// In questo caso, sottoscrive al topic dei dati del sensore.
	opts.SetOnConnectHandler(makeConnectionHandler(filteredDataChannel))

	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		logger.Log.Warn("Sensor lost connection to MQTT broker: ", err.Error())
	})

	client = MQTT.NewClient(opts)

	logger.Log.Info("Sensor attempting to connect to MQTT broker at ", brokerURL)
	if token := client.Connect(); token.WaitTimeout(10*time.Second) && token.Error() != nil {
		//L'errore di connessione inziale viene solo loggato, il meccanismo di
		//AutoReconnect continuerà a tentare in background
		logger.Log.Error("Sensor failed to connect initially:", token.Error())
	}
}

func SetupMQTTConnection(filteredDataChannel chan structure.SensorData) {

	// Assicura che la connessione non sia già stata inizializzata.
	if client != nil && client.IsConnected() {
		logger.Log.Info("MQTT client already connected. Skipping setup.")
		return
	}

	// Inizializza la connessione MQTT
	connectAndManage(filteredDataChannel)

	// Non procedere se la connessione non è attiva.
	if !client.IsConnected() {
		logger.Log.Warn("MQTT client not connected. Exiting setup.")
		os.Exit(1)
	}

	logger.Log.Info("MQTT connection established successfully.")
}
