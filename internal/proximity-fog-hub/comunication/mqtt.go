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

func connectAndManage(filteredDataChannel chan structure.SensorData) {
	if client != nil && client.IsConnected() {
		return
	}

	mqttId := environment.BuildingID + "_" + environment.HubID
	brokerURL := fmt.Sprintf("%s://%s:%s", environment.MosquittoProtocol, environment.MosquittoBroker, environment.MosquittoPort)

	logger.Log.Info("MQTT configuration", "clientID", mqttId, "brokerURL", brokerURL)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(mqttId)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	opts.SetConnectRetry(true)

	// La callback di riconnessione
	opts.SetOnConnectHandler(func(c MQTT.Client) {
		logger.Log.Info("RE-CONNECTED to MQTT broker. Re-subscribing...")
		topic := environment.FilteredDataTopic + "/#"
		token := c.Subscribe(topic, 0, makeSensorDataHandler(filteredDataChannel))
		if token.WaitTimeout(5*time.Second) && token.Error() != nil {
			logger.Log.Error("Failed to re-subscribe to topic", "error", token.Error())
		}
	})

	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		logger.Log.Warn("Lost connection to MQTT broker", "error", err.Error())
	})

	client = MQTT.NewClient(opts)

	logger.Log.Info("Attempting to connect to MQTT broker...")
	if token := client.Connect(); token.WaitTimeout(10*time.Second) && token.Error() != nil {
		logger.Log.Error("Failed to connect to MQTT broker on startup", "error", token.Error())
		os.Exit(1)
	}

	logger.Log.Info("Successfully connected to MQTT broker. Now subscribing...")
	topic := environment.FilteredDataTopic + "/#"
	if token := client.Subscribe(topic, 0, makeSensorDataHandler(filteredDataChannel)); token.WaitTimeout(5*time.Second) && token.Error() != nil {
		logger.Log.Error("Failed to subscribe on initial connection", "topic", topic, "error", token.Error())
		os.Exit(1)
	}

	logger.Log.Info("Successfully subscribed to topic", "topic", topic)
}

func SetupMQTTConnection(filteredDataChannel chan structure.SensorData) {

	// Assicura che la connessione non sia già stata inizializzata.
	if client != nil && client.IsConnected() {
		logger.Log.Info("MQTT client already connected. Skipping setup.")
		return
	}

	// Inizializza la connessione MQTT
	connectAndManage(filteredDataChannel)
}

// Rinominiamo la vecchia funzione, ora si occuperà solo di RI-connessioni
func makeReconnectionHandler(client MQTT.Client, filteredDataChannel chan structure.SensorData) {
	topic := environment.FilteredDataTopic + "/#"
	logger.Log.Info("RE-CONNECTED to MQTT broker. Re-subscribing to topic: ", topic)
	token := client.Subscribe(topic, 0, makeSensorDataHandler(filteredDataChannel))
	if token.WaitTimeout(5*time.Second) && token.Error() != nil {
		logger.Log.Error("Failed to re-subscribe to topic:", topic, "error:", token.Error())
	}
}
