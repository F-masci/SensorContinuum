package comunication

import (
	"SensorContinuum/pkg/structure"
	"encoding/json"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"SensorContinuum/internal/sensor-agent"
	"SensorContinuum/pkg/logger"
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

	mqttId := sensor_agent.BuildingID + "_" + sensor_agent.FloorID + "_" + sensor_agent.SensorID
	opts := MQTT.NewClientOptions().AddBroker(sensor_agent.MosquittoProtocol + "://" + sensor_agent.MosquittoBroker + ":" + sensor_agent.MosquittoPort)
	opts.SetClientID(mqttId)

	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	logger.Log.Info("Connected to MQTT broker")
}

// publish pubblica un messaggio al broker MQTT
func publish(topic string, qos byte, payload string) error {

	// Controlla se il client è connesso, altrimenti stabilisce una connessione
	if client == nil || !client.IsConnected() {
		connect()
	}

	// -------------------------
	// | QoS: 0 (At most once) |
	// | QoS: 1 (At least once)|
	// | QoS: 2 (Exactly once) |
	// -------------------------
	token := client.Publish(topic, qos, false, payload)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	} else {
		logger.Log.Debug("Message published: ", payload)
	}

	return nil

}

// PublishData pubblica i dati del sensore al broker MQTT
func PublishData(data float64) error {

	// Crea la struttura del messaggio
	sensorData := structure.SensorData{
		BuildingID: sensor_agent.BuildingID,
		FloorID:    sensor_agent.FloorID,
		SensorID:   sensor_agent.SensorID,
		Timestamp:  time.Now().Format(time.RFC3339),
		Data:       data,
	}

	// Serializza in JSON
	payload, err := json.Marshal(sensorData)
	if err != nil {
		logger.Log.Error("Error during JSON serialization", err.Error())
		return err
	}

	topic := sensor_agent.BaseTopic + sensor_agent.SensorID
	err = publish(topic, 0, string(payload))
	if err != nil {
		logger.Log.Error("Error during publishing: ", err.Error())
		return err
	}

	return nil

}
