package main

import (
	"SensorContinuum/internal/proximity-fog-hub"
	"SensorContinuum/internal/proximity-fog-hub/aggregation"
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
	"os"
	"time"
)

/*	---------------------------- DESCRIZIONE  PROXIMITY FOG HUB ---------------------------------------------------------

il proximity fog hub svolge i seguenti compiti:

	- riceve i dati filtrati dall edge-hub tramite broker mqtt iscrivendosi allo stesso topic
	- salva i dati ricevuti in una cache locale ( che mantiene i dati delle ultime 6 ore)
	- invia i dati ricevuti tramite kafka all' intermediate-fog-hub, in questo modo l' intermediate-fog-hub
      ha una copia dettagliata e immediata di ogni singolo dato e può rispondere alla domanda " cosa accade ora nel sistema?"
	- ogni 5 minuti (da modificare nel caso) scatta un ticker che:
		1) esegue una query sulla cache locale temporanea (sul db locale quindi) chiedendo di restituire i valori
           di max, min e avg di tutti i dati ricevuti negli ultimi 5 minuti
		2) riceve i risultati dal db
		3) invia queste statistiche all intermediate-fog-hub tramite kafka usando un topic nuovo e dedicato

quindi alla fine avremo una cache locale e saremo sicuri che l 'intermediate-fog-hub possieda una visione riassuntiva e
di più alto livello dello stato dell'edificio

-------------------------------------- FINE DESCRIZIONE ------------------------------------------------------------------------------*/

func main() {
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	logger.CreateLogger(logger.GetProximityHubContext(environment.EdgeMacrozone, environment.HubID))
	logger.PrintCurrentLevel()
	logger.Log.Info("Starting Proximity Fog Hub...")

	// Invia il messaggio di configurazione al Region Hub
	if err := comunication.SendRegistrationMessage(); err != nil {
		logger.Log.Error("Failed to send configuration message to Region Hub, error: ", err)
		os.Exit(1)
	}
	logger.Log.Info("Configuration message sent to Intermediate Fog Hub successfully.")

	// Connessione al DB per la cache
	if err := storage.InitDatabaseConnection(); err != nil {
		logger.Log.Error("failed to connect with local db, error: ", err)
		os.Exit(1)
	}

	// --inizio primo punto di sopra nella descrizione--

	// creazione del canale dove ricevere i dati inviati dall'edge-hub tramite broker MQTT
	filteredDataChannel := make(chan types.SensorData, 100)
	configurationMessageChannel := make(chan types.ConfigurationMsg, 100)
	//connessione e sottoscrizione al topic desiderato del broker MQTT
	comunication.SetupMQTTConnection(filteredDataChannel, configurationMessageChannel)

	// --fine primo punto --

	// -- inizio secondo e terzo punto della descrizione --

	// avvio goroutine ProcessEdgeHubData che salva i dati, ricevuti dall'edge-hub (tramite broker MQTT),
	// nella sua cache locale e li invia tramite kafka all' intermediate-fog-hub
	go proximity_fog_hub.ProcessEdgeHubData(filteredDataChannel)

	go proximity_fog_hub.ProcessEdgeHubConfiguration(configurationMessageChannel)

	// -- fine secondo e terzo punto della descrizione --

	// -- inizio ultimo punto della descrizione --

	// Avvio del ticker per l'aggregazione periodica, per ora metto ogni 2 minuti ma poi passa a ogni 5
	statsTicker := time.NewTicker(2 * time.Minute)
	logger.Log.Info("Ticker started, aggregating data every 2 minutes from now.")
	defer statsTicker.Stop()

	go func() {
		// Esegui subito solo la prima volta, invierà tutti 0
		aggregation.PerformAggregationAndSend()
		// Iniziamo poi con il loop infinito
		for {
			select {
			case <-statsTicker.C:
				aggregation.PerformAggregationAndSend()
			}
		}
	}()

	// -- fine ultimo punto della descrizione--

	logger.Log.Info("Proximity Fog Hub is running. Waiting for termination signal (Ctrl+C)...")
	utils.WaitForTerminationSignal()
	logger.Log.Info("Shutting down Edge Hub...")
}
