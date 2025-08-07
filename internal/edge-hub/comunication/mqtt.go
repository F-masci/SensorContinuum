package comunication

import (
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"encoding/json"
	"fmt"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var client MQTT.Client

// sensorDataHandler è la funzione di callback che processa i messaggi in arrivo.
func makeSensorDataHandler(sensorDataChannel chan structure.SensorData) MQTT.MessageHandler {
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

		case sensorDataChannel <- sensorData:
			// Messaggio inviato correttamente
		default:
			logger.Log.Warn("Data channel is full. Discarding message from sensor: ", sensorData.SensorID)
		}
	}
}

// funzione che viene chiamata quando la connessione MQTT è già riuscita e quello che facciamo ora è
// fare la subscribe al topic dei dati del sensore. Cioè praticamente stiamo dicendo che l edge hub è sottoscritto
// alla ricezione dei dati da parte dei sensori e quindi li riceve
func makeConnectionHandler(sensorDataChannel chan structure.SensorData) MQTT.OnConnectHandler {
	return func(client MQTT.Client) {
		topic := environment.SensorDataTopic + "/#"
		logger.Log.Info("Successfully connected to MQTT broker. Subscribing to topic: ", topic)
		// QoS 0, sottoscrivi al topic per ricevere i dati dai sensori
		token := client.Subscribe(topic, 0, makeSensorDataHandler(sensorDataChannel)) // Il message handler è globale
		logger.Log.Info("Subscribed to topic: ", topic)
		if token.WaitTimeout(5*time.Second) && token.Error() != nil {
			logger.Log.Error("Failed to subscribe to topic:", topic, "error:", token.Error())
			os.Exit(1) // Esci se non riesci a sottoscrivere
		}
	}
}

// connectAndManage gestisce la connessione una sola volta.
func connectAndManage(sensorDataChannel chan structure.SensorData) {
	// Se il client è già definito e connesso, non fare nulla.
	if client != nil && client.IsConnected() {
		return
	}

	mqttId := environment.BuildingID + "_" + environment.FloorID + "_" + environment.HubID
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
	opts.SetOnConnectHandler(makeConnectionHandler(sensorDataChannel))
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

func SetupMQTTConnection(sensorDataChannel chan structure.SensorData) {

	// Assicura che la connessione non sia già stata inizializzata.
	if client != nil && client.IsConnected() {
		logger.Log.Info("MQTT client already connected. Skipping setup.")
		return
	}

	// Inizializza la connessione MQTT
	connectAndManage(sensorDataChannel)

	// Non procedere se la connessione non è attiva.
	if !client.IsConnected() {
		logger.Log.Warn("MQTT client not connected. Exiting setup.")
		os.Exit(1)
	}

	logger.Log.Info("MQTT connection established successfully.")
}

// PublishFilteredData pubblica i dati filtrati al broker MQTT
// Praticamente sopra abbiamo fatto la subscribe nel ricevere i dati dai sensori, ora invece
// facciamo diventare l edge hub un attore che pubblica i dati filtrati ( che saranno presi dal
// proximity fog, il quale farà a sua volta la subscribe al broker mqtt  per rievere i dati
func PublishFilteredData(filteredDataChannel chan structure.SensorData) {

	// Non procedere se la connessione non è attiva.
	if !client.IsConnected() {
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
		// invia i dati al broker MQTT
		token := client.Publish(topic, 0, false, payload)

		// Usiamo WaitTimeout per non bloccare il sensore all'infinito,
		// cioè se la rete è lenta il sensore comunque non si blocca
		//anche se il timeout scade il programma comunque prosegue
		if !token.WaitTimeout(2 * time.Second) {
			logger.Log.Warn("Timeout publishing message.")
		} else if err := token.Error(); err != nil {
			logger.Log.Error("Error publishing message: ", err.Error())
		} else {
			logger.Log.Debug("Message published successfully.")
		}
	}
}
