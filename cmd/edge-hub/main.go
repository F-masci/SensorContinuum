package main

import (
	edge_hub "SensorContinuum/internal/edge-hub"
	"SensorContinuum/internal/edge-hub/comunication"
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/internal/edge-hub/health"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
	"math/rand"
	"os"
	"time"
)

/**
 * Punto di ingresso dell'applicazione Edge Hub.
 * Inizializza l'ambiente, il logger, i canali di comunicazione e
 * avvia i servizi in base alla modalità di servizio configurata.
 * Gestisce i servizi di configurazione, filtro, aggregazione e pulizia.
 * Avvia anche il server di health check se abilitato.
 */
func main() {

	// Setup dell'ambiente
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	// Inizializza il logger
	logger.CreateLogger(logger.GetEdgeHubContext(environment.ServiceMode, environment.EdgeMacrozone, environment.EdgeZone, environment.HubID))
	logger.PrintCurrentLevel()
	logger.Log.Info("Starting Edge Hub...")
	logger.Log.Info("Hub service mode: ", environment.ServiceMode)

	// Creazione del canale per i messaggi di configurazione
	sensorConfigurationMessageChannel := make(chan types.ConfigurationMsg, 200)
	// creazione del canale per i dati ricevuti dai sensori
	sensorDataChannel := make(chan types.SensorData, 200)
	// inizializza connessione MQTT in maniera sincrona
	comunication.SetupMQTTConnection(sensorDataChannel, sensorConfigurationMessageChannel)

	// Si registra al proximity Hub in base al proprio Service Mode
	// Questo invio è sincrono, se fallisce l'applicazione termina
	logger.Log.Info("Sending registration message")
	comunication.SendRegistrationMessage()
	logger.Log.Info("Registration message sent successfully")

	// Avvia il thread per l'invio dei messaggi di heartbeat
	go comunication.SendHeartbeatMessage()

	/* ----- CONFIGURATION SERVICE ------ */

	if environment.ServiceMode == types.EdgeHubConfigurationService || environment.ServiceMode == types.EdgeHubService {

		// Avvia l'elaborazione dei messaggi di configurazione in un'altra goroutine.
		hubConfigurationMessageChannel := make(chan types.ConfigurationMsg, 200)
		go edge_hub.ProcessSensorConfigurationMessages(sensorConfigurationMessageChannel, hubConfigurationMessageChannel)
		go comunication.PublishConfigurationMessage(hubConfigurationMessageChannel)

	}

	/* ----- FILTER SERVICE ------ */

	if environment.ServiceMode == types.EdgeHubFilterService || environment.ServiceMode == types.EdgeHubService {

		// Avvia il filtro in un'altra goroutine.
		go edge_hub.FilterSensorData(sensorDataChannel)

	}

	/* ----- AGGREGATOR SERVICE ------ */

	if (environment.ServiceMode == types.EdgeHubAggregatorService && environment.OperationMode == types.OperationModeLoop) || environment.ServiceMode == types.EdgeHubService {

		// Creazione del canale per i dati filtrati
		filteredDataChannel := make(chan types.SensorData, 200)
		// Aspettiamo che arrivino i dati sul canale filteredDataChannel e li invia via MQTT
		go comunication.PublishFilteredData(filteredDataChannel)

		// Crea un ticker che scatta ogni AggregationInterval (1 minuto di default).
		// Ogni volta che scatta, chiama AggregateAllSensorsData per aggregare i dati.
		// Per evitare che tutti gli edge hub facciano l'aggregazione nello stesso istante,
		// si calcola un ritardo casuale tra 0 e AggregationInterval all'inizio.
		// In questo modo, anche se tutti gli edge hub partono nello stesso momento,
		// l'aggregazione avverrà in momenti diversi, cercando di evitare race condition.
		// Calcola il prossimo tick allineato all'intervallo
		logger.Log.Debug("Calculating initial random delay for aggregation ticker")
		now := time.Now()
		interval := environment.AggregationInterval
		startOfInterval := now.Truncate(interval)
		secondsSinceStart := now.Sub(startOfInterval).Seconds()
		secondsToEnd := interval.Seconds() - secondsSinceStart

		applyDelay := secondsSinceStart <= 5 || secondsToEnd <= 5

		var randomDelay time.Duration
		if applyDelay {
			// Genera un offset casuale tra 5 e interval-5
			offset := time.Duration(5 + rand.Int63n(int64(interval)-5))
			nextTick := startOfInterval.Add(interval)
			randomDelay = nextTick.Sub(now) + offset
			logger.Log.Info("Applying initial random delay of ", randomDelay.String())
			time.Sleep(randomDelay)
		} else {
			logger.Log.Info("No initial random delay applied")
		}

		aggregateTicker := time.NewTicker(interval)
		defer aggregateTicker.Stop()
		logger.Log.Info("Aggregation ticker started with interval ", interval.String())

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

	if environment.ServiceMode == types.EdgeHubAggregatorService && environment.OperationMode == types.OperationModeOnce {

		// Creazione del canale per i dati filtrati
		filteredDataChannel := make(chan types.SensorData, 200)
		// Aspettiamo che arrivino i dati sul canale filteredDataChannel e li invia via MQTT
		go comunication.PublishFilteredData(filteredDataChannel)

		// Esegue una singola aggregazione e termina
		edge_hub.AggregateAllSensorsData(filteredDataChannel)
		logger.Log.Info("Aggregation completed. The service will now terminate.")
		os.Exit(0)
	}

	/* ----- CLEAN SERVICE ------ */

	// duale a sopra solo rivolto al caso del processo di clean.

	if (environment.ServiceMode == types.EdgeHubCleanerService && environment.OperationMode == types.OperationModeLoop) || environment.ServiceMode == types.EdgeHubService {

		cleanHealthTicker := time.NewTicker(time.Minute)
		defer cleanHealthTicker.Stop()
		logger.Log.Info("Starting cleaning ticker with interval 1 minute")

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

	if environment.ServiceMode == types.EdgeHubCleanerService && environment.OperationMode == types.OperationModeOnce {
		unhealthySensors, removedSensors := edge_hub.CleanUnhealthySensors()
		edge_hub.NotifyUnhealthySensors(unhealthySensors)
		edge_hub.NotifyRemovedSensors(removedSensors)
		logger.Log.Info("Cleaning completed. The service will now terminate.")
		os.Exit(0)
	}

	/* -------- HEALTH CHECK SERVER -------- */

	// Avvia il server di health check se abilitato
	if environment.HealthzServer {
		logger.Log.Info("Enabling health check channel on port " + environment.HealthzServerPort)
		go func() {
			if err := health.StartHealthCheckServer(":" + environment.HealthzServerPort); err != nil {
				logger.Log.Error("Failed to enable health check channel: ", err.Error())
				os.Exit(1)
			}
		}()
	}

	// Attende il segnale di terminazione (ad esempio Ctrl+C)
	utils.WaitForTerminationSignal()
	logger.Log.Info("Shutting down Edge Hub...")
}
