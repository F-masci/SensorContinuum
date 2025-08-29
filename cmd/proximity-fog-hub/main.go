package main

import (
	"SensorContinuum/internal/proximity-fog-hub"
	"SensorContinuum/internal/proximity-fog-hub/aggregation"
	"SensorContinuum/internal/proximity-fog-hub/cleaner"
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/dispatcher"
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/internal/proximity-fog-hub/health"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
	"context"
	"os"
	"time"
)

/*	---------------------------- DESCRIZIONE  PROXIMITY FOG HUB ---------------------------------------------------------

il proximity fog hub svolge i seguenti compiti:

	- riceve i dati filtrati dall edge-hub tramite broker mqtt iscrivendosi allo stesso topic
	- salva i dati ricevuti in una cache locale (che mantiene i dati delle ultime 6 ore)
	- invia i dati ricevuti tramite kafka all' intermediate-fog-hub, in questo modo l' intermediate-fog-hub
      ha una copia dettagliata e immediata di ogni singolo dato e può rispondere alla domanda "cosa accade ora nel sistema?"
	- ogni 5 minuti (da modificare nel caso) scatta un ticker che:
		1) esegue una query sulla cache locale temporanea (sul db locale quindi) chiedendo di restituire i valori
            di max, min e avg di tutti i dati ricevuti negli ultimi 5 minuti
		2) riceve i risultati dal db
		3) salva queste statistiche in una tabella outbox locale, ci penserà poi un componente chiamato
			dispatcher (una goroutine che si avvia anche essa grazie a un ticker)
			a inviarle all' intermediate-fog-hub e settare lo stato 'sent' alla tabella outbox.
		4) un cleaner che si avvia sempre con un ticker in una goroutine separata si occupa di pulire la tabella outbox
			ossia nello specifico di eliminare i messaggi che sono in stato 'sent' e che hanno più di 1 ora (per il momento)
			per evitare che la tabella outbox cresca all'infinito

quindi alla fine saremo sicuri che l 'intermediate-fog-hub possieda una visione riassuntiva e
di più alto livello dello stato dell'edificio

------------------------------------------------------------------------------------------------------------------------- */

func main() {
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	logger.CreateLogger(logger.GetProximityHubContext(environment.EdgeMacrozone, environment.HubID))
	logger.PrintCurrentLevel()
	logger.Log.Info("Starting Proximity Fog Hub...")

	// Invia il messaggio di configurazione al Region Hub
	if err := proximity_fog_hub.SendOwnRegistrationMessage(); err != nil {
		logger.Log.Error("Failed to send configuration message to Region Hub, error: ", err)
		os.Exit(1)
	}
	logger.Log.Info("Configuration message sent to Intermediate Fog Hub successfully.")

	// Invia i propri messaggi di heartbeat per segnalare che il Proximity Fog Hub è attivo
	go proximity_fog_hub.SendOwnHeartbeatMessage()

	// Connessione al DB per la cache
	if err := storage.InitDatabaseConnection(); err != nil {
		logger.Log.Error("failed to connect with local db, error: ", err)
		os.Exit(1)
	}

	// --inizio primo punto di sopra nella descrizione--

	// creazione del canale dove ricevere i dati inviati dall' edge-hub tramite broker MQTT
	filteredDataChannel := make(chan types.SensorData, 100)
	configurationMessageChannel := make(chan types.ConfigurationMsg, 100)
	heartbeatMessageChannel := make(chan types.HeartbeatMsg, 100)
	//connessione e sottoscrizione al topic desiderato del broker MQTT
	comunication.SetupMQTTConnection(filteredDataChannel, configurationMessageChannel, heartbeatMessageChannel)

	// --fine primo punto --

	// -- inizio secondo e terzo punto della descrizione --

	// avvio goroutine ProcessEdgeHubData che salva i dati, ricevuti dall' edge-hub (tramite broker MQTT),
	// nella sua cache locale e li invia tramite kafka all' intermediate-fog-hub
	go proximity_fog_hub.ProcessEdgeHubData(filteredDataChannel)

	go proximity_fog_hub.ProcessEdgeHubConfiguration(configurationMessageChannel)

	// Avvio del ticker per l'aggregazione periodica, per ora metto ogni 2 minuti ma poi passa a ogni 5
	statsTicker := time.NewTicker(5 * time.Minute)
	logger.Log.Info("Ticker started, aggregating data every 2 minutes from now.")
	defer statsTicker.Stop()

	go func() {
		// Esegui subito solo la prima volta, invierà tutti 0
		//aggregation.PerformAggregationAndSend()
		// Iniziamo poi con il loop infinito
		for {
			select {
			case <-statsTicker.C:
				aggregation.PerformAggregationAndSend()
			}
		}
	}()

	// avvio del dispatcher outbox in una goroutine separata, questo processo
	// si occupa di inviare i dati aggregati che sono stati salvati nella tabella outbox locale
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go dispatcher.Run(ctx)

	// -- fine secondo e terzo punto della descrizione --

	// -- inizio ultimo punto della descrizione --

	// Avviamo il cleaner in una goroutine separata.
	// Questo processo si occuperà di pulire periodicamente la tabella
	// outbox, rimuovendo i messaggi già inviati da tempo.
	go cleaner.Run(ctx)

	// -- fine ultimo punto della descrizione--

	// Processo i messaggi di heartbeat ricevuti dall' edge-hub
	go proximity_fog_hub.ProcessEdgeHubHeartbeat(heartbeatMessageChannel)

	/* -------- HEALTH CHECK SERVER -------- */

	if environment.HealthzServer {
		logger.Log.Info("Enabling health check channel on port " + environment.HealthzServerPort)
		go func() {
			if err := health.StartHealthCheckServer(":" + environment.HealthzServerPort); err != nil {
				logger.Log.Error("Failed to enable health check channel: ", err.Error())
				os.Exit(1)
			}
		}()
	}

	logger.Log.Info("Proximity Fog Hub is running. Waiting for termination signal (Ctrl+C)...")
	// Attende il segnale di terminazione (ad esempio Ctrl+C)
	utils.WaitForTerminationSignal()
	logger.Log.Info("Shutting down Edge Hub...")
}
