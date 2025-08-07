package main

import (
	"SensorContinuum/internal/edge-hub"
	"SensorContinuum/internal/edge-hub/comunication"
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"SensorContinuum/pkg/utils"
	"os"
	"time"
)

func main() {
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}
	const service = "edge-hub"
	logger.CreateLogger(logger.GetContext(service, environment.BuildingID, environment.FloorID, environment.HubID))
	logger.Log.Info("Starting Edge Hub...")

	//creazione del canale per i dati ricevuti dai sensori
	sensorDataChannel := make(chan structure.SensorData, 100)
	// inizializza connessione MQTT in maniera sincrona
	comunication.SetupMQTTConnection(sensorDataChannel)

	// Avvia il filtro in un'altra goroutine.
	go edge_hub.FilterSensorData(sensorDataChannel)

	// creazione del canale per i dati filtrati
	filteredDataChannel := make(chan structure.SensorData, 100)
	// Aspettiamo che arrivino i dati sul canale filterdDataChannel e li invia via MQTT
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

	// duale a sopra solo rivolto al caso del processo di clean.
	cleanHealthTicker := time.NewTicker(time.Minute)
	defer cleanHealthTicker.Stop()

	go func() {
		for {
			select {
			case <-cleanHealthTicker.C:
				unhealthySensors := edge_hub.CleanUnhealthySensors()
				edge_hub.NotifyUnhealthySensors(unhealthySensors)
			}
		}
	}()

	utils.WaitForTerminationSignal()

	logger.Log.Info("Edge Hub is terminating")
	logger.Log.Info("Shutting down Edge Hub...")
}
