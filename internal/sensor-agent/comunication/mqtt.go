package comunication

import (
	"SensorContinuum/internal/sensor-agent/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"encoding/json"
	"fmt"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var client MQTT.Client

// connectAndManage gestisce la connessione una sola volta.
func connectAndManage() {
	// Se il client è già definito e connesso, non fare nulla.
	if client != nil && client.IsConnected() {
		return
	}

	mqttId := environment.BuildingID + "_" + environment.FloorID + "_" + environment.SensorID
	brokerURL := fmt.Sprintf("%s://%s:%s", environment.MosquittoProtocol, environment.MosquittoBroker, environment.MosquittoPort)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(mqttId)

	// --- Impostazioni di Resilienza ---

	// la libreria paho gestisce automaticamente la riconnessione in background,
	opts.SetAutoReconnect(true)
	//controlla la frequenza di tenta della riconnessione
	opts.SetMaxReconnectInterval(10 * time.Second)
	opts.SetConnectRetry(true)

	opts.SetOnConnectHandler(func(c MQTT.Client) {
		logger.Log.Info("Sensor connected to MQTT broker.")
	})
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

// PublishData pubblica i dati del sensore al broker MQTT
func PublishData(data float64) {
	// Assicura che la connessione sia gestita
	if client == nil {
		connectAndManage()
	}

	// Non procedere se la connessione non è attiva.
	if !client.IsConnected() {
		logger.Log.Warn("MQTT client not connected. Skipping data publishing.")
		// L'opzione AutoReconnect della libreria sta già lavorando per riconnettersi.
		return
	}

	sensorData := structure.SensorData{
		BuildingID: environment.BuildingID,
		FloorID:    environment.FloorID,
		SensorID:   environment.SensorID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Data:       data,
	}

	payload, err := json.Marshal(sensorData)
	if err != nil {
		logger.Log.Error("Error during JSON serialization: ", err.Error())
		return
	}

	topic := environment.BaseTopic + environment.SensorID
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
