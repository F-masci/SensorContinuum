package comunication

import (
	"SensorContinuum/internal/edge-hub"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var client MQTT.Client
var connectionPending = false

// connect stabilisce una connessione al broker MQTT
func connect() {

	if connectionPending {
		logger.Log.Warn("Connection already in progress, skipping new connection attempt")
		return
	}

	connectionPending = true

	mqttId := edge_hub.BuildingID + "_" + edge_hub.FloorID + "_" + edge_hub.HubID
	opts := MQTT.NewClientOptions().AddBroker(edge_hub.MosquittoProtocol + "://" + edge_hub.MosquittoBroker + ":" + edge_hub.MosquittoPort)
	opts.SetClientID(mqttId)

	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	logger.Log.Info("Connected to MQTT broker")
}

// subscribe sottoscrive un topic specifico con il QoS desiderato
func subscribe(topic string, qos byte, callback func(client MQTT.Client, msg MQTT.Message)) error {

	// Controlla se il client è connesso, altrimenti stabilisce una connessione
	if client == nil || !client.IsConnected() {
		connect()
	}

	// Sottoscrive al topic specificato con il QoS desiderato
	// -------------------------
	// | QoS: 0 (At most once) |
	// | QoS: 1 (At least once)|
	// | QoS: 2 (Exactly once) |
	// -------------------------
	token := client.Subscribe(topic, qos, callback)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	logger.Log.Info("Subscribed to topic: ", topic)
	return nil
}

// PullSensorData sottoscrive al topic dei dati dei sensori e invia i dati ricevuti al canale specificato
func PullSensorData(dataChannel chan structure.SensorData) error {

	topic := edge_hub.BaseTopic + "#"
	err := subscribe(topic, 0, func(client MQTT.Client, msg MQTT.Message) {

		logger.Log.Debug("Received message on topic: ", msg.Topic(), " with payload: ", string(msg.Payload()))

		// Simula l'estrazione dei dati dal messaggio ricevuto
		sensorData, err := structure.CreateSensorDataFromMQTT(msg)
		if err != nil {
			logger.Log.Error("Error parsing sensor data from MQTT message: ", err.Error())
		} else {
			dataChannel <- sensorData
		}
	})

	return err

}
