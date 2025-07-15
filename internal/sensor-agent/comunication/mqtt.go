package comunication

import (
	"fmt"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	mosquittoConf "SensorContinuum/configs/mosquitto"
	"SensorContinuum/internal/sensor-agent"
	"SensorContinuum/pkg/logger"
)

var client MQTT.Client
var connectionPending = false

func connect() {

	if connectionPending {
		logger.Log.Warn("Connection already in progress, skipping new connection attempt")
		return
	}

	connectionPending = true

	address, exists := os.LookupEnv("MQTT_BROKER")
	if !exists {
		address = "localhost"
	}

	opts := MQTT.NewClientOptions().AddBroker(mosquittoConf.PROTOCOL + "://" + address + ":" + mosquittoConf.PORT)
	opts.SetClientID(sensor_agent.GetSensorID())

	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	logger.Log.Info("Connected to MQTT broker")
}

func Publish(data float64) {

	if client == nil || !client.IsConnected() {
		connect()
	}

	token := client.Publish("floor/1/sensors/"+sensor_agent.GetSensorID(), 0, false, fmt.Sprintf("%f", data))
	token.Wait()
	if token.Error() != nil {
		logger.Log.Error("Error during publishing: ", token.Error())
	} else {
		logger.Log.Debug("Message published: ", data)
	}
}
