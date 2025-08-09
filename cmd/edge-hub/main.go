package main

import (
	"SensorContinuum/internal/edge-hub"
	"SensorContinuum/internal/edge-hub/comunication"
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
	"os"
	"time"
)

func main() {

	// Inizializza l'ambiente e il logger
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	logger.CreateLogger(logger.GetEdgeHubContext(environment.ServiceMode, environment.EdgeMacrozone, environment.EdgeZone, environment.HubID))
	logger.PrintCurrentLevel()
	logger.Log.Info("Starting Edge Hub...")
	logger.Log.Info("Hub service mode: ", environment.ServiceMode)

	// creazione del canale per i messaggi di configurazione
	sensorConfigurationMessageChannel := make(chan types.ConfigurationMsg, 100)
	// creazione del canale per i dati ricevuti dai sensori
	sensorDataChannel := make(chan types.SensorData, 200)
	// inizializza connessione MQTT in maniera sincrona
	comunication.SetupMQTTConnection(sensorDataChannel, sensorConfigurationMessageChannel)

	// Si registra al proximity Hub come servizio
	logger.Log.Info("Sending registration message")
	comunication.SendRegistrationMessage()
	logger.Log.Info("Registration message sent successfully")

	/* ----- CONFIGURATION SERVICE ------ */

	if environment.ServiceMode == types.EdgeHubConfigurationService || environment.ServiceMode == types.EdgeHubService {

		// Avvia l'elaborazione dei messaggi di configurazione in un'altra goroutine.
		hubConfigurationMessageChannel := make(chan types.ConfigurationMsg, 100)
		go edge_hub.ProcessSensorConfigurationMessages(sensorConfigurationMessageChannel, hubConfigurationMessageChannel)
		go comunication.PublishConfigurationMessage(hubConfigurationMessageChannel)

	}

	/* ----- FILTER SERVICE ------ */

	if environment.ServiceMode == types.EdgeHubFilterService || environment.ServiceMode == types.EdgeHubService {

		// Avvia il filtro in un'altra goroutine.
		go edge_hub.FilterSensorData(sensorDataChannel)

	}

	/* ----- AGGREGATION SERVICE ------ */

	if (environment.ServiceMode == types.EdgeHubAggregatorService && environment.OperationMode == environment.OperationModeLoop) || environment.ServiceMode == types.EdgeHubService {

		// creazione del canale per i dati filtrati
		filteredDataChannel := make(chan types.SensorData, 100)
		// Aspettiamo che arrivino i dati sul canale filteredDataChannel e li invia via MQTT
		go comunication.PublishFilteredData(filteredDataChannel)

		// creazione di un timer per i dati aggregati che invia un segnale, ogni minuto, sul suo canale "C".
		aggregateTicker := time.NewTicker(time.Minute)
		defer aggregateTicker.Stop()

		// avvia una goroutine che vivrà per sempre
		go func() {
			//loop infinito
			for {
				// mettiti in attesa
				select {
				//il codice si blocca aspettando che il ticker invii il segnale (ogni minuto)
				// quando arriva il segnale, viene chiamata AggregateAllSensorsData per l'aggregazione dei dati filtrati.
				case <-aggregateTicker.C:
					edge_hub.AggregateAllSensorsData(filteredDataChannel)
				}
			}
		}()

	}

	if environment.ServiceMode == types.EdgeHubAggregatorService && environment.OperationMode == environment.OperationModeOnce {

		filteredDataChannel := make(chan types.SensorData, 100)
		go comunication.PublishFilteredData(filteredDataChannel)

		edge_hub.AggregateAllSensorsData(filteredDataChannel)
		logger.Log.Info("Aggregation completed. The service will now terminate.")
		os.Exit(0)
	}

	/* ----- CLEAN SERVICE ------ */

	// duale a sopra solo rivolto al caso del processo di clean.

	if (environment.ServiceMode == types.EdgeHubCleanerService && environment.OperationMode == environment.OperationModeLoop) || environment.ServiceMode == types.EdgeHubService {

		cleanHealthTicker := time.NewTicker(time.Minute)
		defer cleanHealthTicker.Stop()

		go func() {
			for {
				select {
				case <-cleanHealthTicker.C:
					unhealthySensors, removedSensors := edge_hub.CleanUnhealthySensors()
					edge_hub.NotifyUnhealthySensors(unhealthySensors)
					edge_hub.NotifyRemovedSensors(removedSensors)
				}
			}
		}()

	}

	if environment.ServiceMode == types.EdgeHubCleanerService && environment.OperationMode == environment.OperationModeOnce {
		unhealthySensors, removedSensors := edge_hub.CleanUnhealthySensors()
		edge_hub.NotifyUnhealthySensors(unhealthySensors)
		edge_hub.NotifyRemovedSensors(removedSensors)
		logger.Log.Info("Cleaning completed. The service will now terminate.")
		os.Exit(0)
	}

	utils.WaitForTerminationSignal()

	logger.Log.Info("Edge Hub is terminating")
	logger.Log.Info("Shutting down Edge Hub...")
}
