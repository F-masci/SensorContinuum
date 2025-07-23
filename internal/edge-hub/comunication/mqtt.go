package comunication

import (
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"errors"
	"fmt"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var client MQTT.Client

// messageHandler è la funzione di callback che processa i messaggi in arrivo.
func messageHandler(dataChannel chan structure.SensorData) MQTT.MessageHandler {
	return func(client MQTT.Client, msg MQTT.Message) {
		logger.Log.Debug("Received message on topic: ", msg.Topic())

		sensorData, err := structure.CreateSensorDataFromMQTT(msg)
		if err != nil {
			logger.Log.Error("Error parsing sensor data from MQTT message: ", err.Error())
			return
		}

		// select non bloccante , il ricevitore MQTT non viene mai bloccato dal processore
		// Se il canale è pieno, scarta il messaggio per non rallentare la ricezione.
		select {
		case dataChannel <- sensorData:
			// Messaggio inviato correttamente
		default:
			logger.Log.Warn("Data channel is full. Discarding message from sensor: ", sensorData.SensorID)
		}
	}
}

// onConnectHandler viene eseguito ogni volta che la connessione al broker ha successo.
func onConnectHandler(client MQTT.Client) {
	topic := environment.SensorDataTopic + "#"
	logger.Log.Info("Successfully connected to MQTT broker. Subscribing to topic...")
	// QoS 0, sottoscrivi al topic per ricevere i dati
	token := client.Subscribe(topic, 0, nil) // Il message handler è globale
	if token.WaitTimeout(5*time.Second) && token.Error() != nil {
		logger.Log.Error("Failed to subscribe to topic:", topic, "error:", token.Error())
	} else {
		logger.Log.Info("Successfully subscribed to topic:", topic)
		// Create a file to indicate health
		if f, err := os.Create("/tmp/healthy"); err != nil {
			logger.Log.Error("failed to create health check file", "error", err)
		} else {
			f.Close()
		}
	}
}

// PullSensorData inizializza e gestisce la connessione MQTT in modo robusto.
// Non ritorna mai, mantenendo viva la connessione.
func PullSensorData(dataChannel chan structure.SensorData) error {
	mqttId := environment.BuildingID + "_" + environment.FloorID + "_" + environment.HubID
	brokerURL := fmt.Sprintf("%s://%s:%s", environment.MosquittoProtocol, environment.MosquittoBroker, environment.MosquittoPort)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(mqttId)
	opts.SetProtocolVersion(5)

	// --- Impostazioni di Resilienza ---
	opts.SetAutoReconnect(true)                    // Abilita la riconnessione automatica
	opts.SetMaxReconnectInterval(10 * time.Second) // Intervallo massimo tra i tentativi
	opts.SetConnectRetry(true)                     // Riprova la connessione all'avvio se fallisce

	// Definisci i callback per gestire eventi di connessione
	opts.SetDefaultPublishHandler(messageHandler(dataChannel)) // Handler per i messaggi
	opts.SetOnConnectHandler(onConnectHandler)                 // Eseguito quando la connessione è (ri)stabilita
	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		logger.Log.Warn("Connection lost with MQTT broker: ", err.Error())
		os.Remove("/tmp/healthy")
	})

	client = MQTT.NewClient(opts)

	logger.Log.Info("Attempting to connect to MQTT broker at ", brokerURL)
	if token := client.Connect(); token.WaitTimeout(10*time.Second) && token.Error() != nil {
		// Non usiamo panic, ritorniamo un errore. La main deciderà cosa fare.
		logger.Log.Error("Fatal error connecting to MQTT broker:", token.Error())
		return errors.New("could not connect to MQTT broker after initial attempt")
	}

	// Blocca la goroutine per sempre. La libreria Paho gestirà tutto in background.
	select {}
}
