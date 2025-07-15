package comunication

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"os"

	"SensorContinuum/pkg/logger"
)

var client MQTT.Client

func connect() {

	adress, exists := os.LookupEnv("MQTT_BROKER")
	if !exists {
		adress = "localhost"
	}

	opts := MQTT.NewClientOptions().AddBroker("tcp://" + adress + ":1883")
	opts.SetClientID("go_mqtt_client")

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

	token := client.Publish("piano/1/sensori/temperatura/sensore01", 0, false, fmt.Sprintf("%f", data))
	token.Wait()
	if token.Error() != nil {
		logger.Log.Error("Error during publishing: ", token.Error())
	} else {
		logger.Log.Debug("Message published: ", data)
	}
}
