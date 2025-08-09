package comunication

import (
	"SensorContinuum/internal/sensor-agent/environment"
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

var client MQTT.Client

// connectAndManage gestisce la connessione una sola volta.
func connectAndManage() {
	// Se il client è già definito e connesso, non fare nulla.
	if client != nil && client.IsConnected() {
		return
	}

	mqttId := environment.EdgeMacrozone + "_" + environment.EdgeZone + "_" + environment.SensorId
	logger.Log.Debug(fmt.Sprintf("MQTT Client ID: %s", mqttId))
	brokerURL := fmt.Sprintf("%s://%s:%s", environment.MqttBrokerProtocol, environment.MqttBrokerAddress, environment.MqttBrokerPort)
	logger.Log.Debug(fmt.Sprintf("Broker URL: %s", brokerURL))

	opts := MQTT.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(mqttId)

	// --- Impostazioni di Resilienza ---

	// la libreria paho gestisce automaticamente la riconnessione in background,
	opts.SetAutoReconnect(true)
	// controlla la frequenza di tenta della riconnessione
	logger.Log.Debug("Setting MQTT Max Reconnect Interval to ", environment.MaxReconnectionInterval, " seconds")
	opts.SetMaxReconnectInterval(time.Duration(environment.MaxReconnectionInterval) * time.Second)
	opts.SetConnectRetry(true)

	var connectAttempts int

	opts.SetOnConnectHandler(func(c MQTT.Client) {
		logger.Log.Info("Sensor connected to MQTT broker.")
		connectAttempts = 0 // resetta i tentativi dopo una connessione riuscita
	})
	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		logger.Log.Warn("Sensor lost connection to MQTT broker: ", err.Error())
	})

	// Limita il numero di tentativi di connessione
	// Se il numero di tentativi supera maxConnectAttempts, il programma termina.
	//
	// Questo è utile per evitare loop infiniti in caso di problemi di connessione persistenti
	// e per evitare che il sensore continui a tentare di connettersi
	// in un ciclo infinito senza successo.
	//
	// Il rischio è che il sensore si riconnetta dopo che il suo messaggio di registrazione
	// è stato cancellato. In questo modo, forzando il riavvio del sensore, si evita
	// che il sensore comunichi dopo che il suo messaggio di registrazione è stato cancellato.
	opts.SetConnectionAttemptHandler(func(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
		connectAttempts++
		if connectAttempts > environment.MaxReconnectionAttempts {
			logger.Log.Error("Max connection attempts reached. Exiting.")
			os.Exit(1)
		}
		logger.Log.Debug("Sensor attempting to connect to MQTT broker: ", connectAttempts, " attempt(s) on ", environment.MaxReconnectionAttempts, " max attempt(s)")
		return tlsCfg
	})

	client = MQTT.NewClient(opts)

	logger.Log.Info("Sensor attempting to connect to MQTT broker at ", brokerURL)
	if token := client.Connect(); token.WaitTimeout(time.Duration(environment.MaxReconnectionTimeout)*time.Second) && token.Error() != nil {
		// L'errore di connessione iniziale viene solo loggato, il meccanismo di
		// AutoReconnect continuerà a tentare in background
		logger.Log.Error("Sensor failed to connect:", token.Error())
	}
}

// PublishData pubblica i dati del sensore al broker MQTT
func PublishData(sensorChannel chan types.SensorData) {

	// Assicura che la connessione sia gestita
	if client == nil {
		connectAndManage()
	}

	for sensorData := range sensorChannel {

		// Non procedere se la connessione non è attiva.
		if !client.IsConnected() {
			logger.Log.Warn("MQTT client not connected. Skipping data publishing.")
			// L'opzione AutoReconnect della libreria sta già lavorando per riconnettersi.
			// Quindi non è necessario riconnettersi manualmente qui.
			// Il sensore continuerà a tentare di riconnettersi in background.
			time.Sleep(time.Duration(environment.MaxReconnectionInterval) * time.Second)
			continue
		}

		payload, err := json.Marshal(sensorData)
		if err != nil {
			logger.Log.Warn("Error during JSON serialization: ", err.Error(), ". Skipping data publishing.")
			continue
		}

		topic := environment.DataTopic + "/" + environment.SensorId
		token := client.Publish(topic, 0, false, payload)

		// Usiamo WaitTimeout per non bloccare il sensore all'infinito,
		// cioè se la rete è lenta il sensore comunque non si blocca
		// anche se il timeout scade il programma comunque prosegue
		if !token.WaitTimeout(time.Duration(environment.MessagePublishTimeout) * time.Second) {
			logger.Log.Warn("Timeout publishing message (", environment.MessagePublishTimeout, " seconds) to MQTT broker.")
		} else if err := token.Error(); err != nil {
			logger.Log.Error("Error publishing message: ", err.Error())
		} else {
			logger.Log.Debug("Message published successfully on topic: ", topic)
		}
	}
}

// SendRegistrationMessage invia un messaggio di registrazione al broker MQTT
// per registrare il sensore con le sue informazioni di configurazione.
func SendRegistrationMessage() {

	// Assicura che la connessione sia gestita
	if client == nil {
		connectAndManage()
	}

	// Non procedere se il messaggio di configurazione non viene inviato
	for {

		// Non procedere se la connessione non è attiva.
		if !client.IsConnected() {
			logger.Log.Warn("MQTT client not connected. Skipping data publishing.")
			// L'opzione AutoReconnect della libreria sta già lavorando per riconnettersi.
			return
		}

		payload, err := json.Marshal(types.ConfigurationMsg{
			EdgeMacrozone:   environment.EdgeMacrozone,
			MsgType:         types.NewSensorMsgType,
			Timestamp:       time.Now().Unix(),
			Service:         types.SensorAgentService,
			EdgeZone:        environment.EdgeZone,
			SensorID:        environment.SensorId,
			SensorLocation:  string(environment.SensorLocation),
			SensorType:      string(environment.SensorType),
			SensorReference: string(environment.SimulationSensorReference),
		})
		if err != nil {
			logger.Log.Error("Error during JSON serialization: ", err.Error())
			os.Exit(1)
		}

		// QoS (Quality of Service) in MQTT:
		// 0: At most once - Nessuna conferma, il messaggio può andare perso.
		// 1: At least once - Il messaggio viene consegnato almeno una volta, può essere duplicato.
		// 2: Exactly once - Il messaggio viene consegnato una sola volta, senza duplicati.
		//
		// Retained:
		// true  - Il broker conserva l’ultimo messaggio pubblicato su un topic e lo invia ai nuovi iscritti.
		// false - Il messaggio non viene conservato dal broker.
		topic := environment.ConfigurationTopic + "/" + environment.SensorId
		token := client.Publish(topic, 0, true, payload)

		// Usiamo WaitTimeout per non ciclare all'infinito
		if !token.WaitTimeout(time.Duration(environment.MessagePublishTimeout) * time.Second) {
			logger.Log.Warn("Timeout publishing configuration message (", environment.MessagePublishTimeout, " seconds) to MQTT broker. Retryng...")
			continue
		} else if err := token.Error(); err != nil {
			logger.Log.Error("Error publishing message: ", err.Error())
			os.Exit(1)
		} else {
			logger.Log.Debug("Message published successfully on topic: ", topic)
			return
		}

	}
}

func IsConnected() bool {

	// Assicura che la connessione sia gestita
	if client == nil {
		connectAndManage()
	}

	// Controlla se il client è connesso
	if client.IsConnected() {
		logger.Log.Debug("MQTT client is connected.")
		return true
	} else {
		logger.Log.Warn("MQTT client is not connected.")
		return false
	}
}
