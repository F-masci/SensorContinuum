package main

import (
	"SensorContinuum/internal/sensor-agent/comunication"
	"SensorContinuum/internal/sensor-agent/environment"
	"SensorContinuum/internal/sensor-agent/health"
	"SensorContinuum/internal/sensor-agent/simulation"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
	"os"
)

/*
DESCRIZIONE FUNZIONALE:
Il Sensor Agent opera come il nodo computazionale di base nel Compute Continuum, simulando il comportamento di un sensore IoT sul campo.

RESPONSABILITÀ CHIAVE:

1.  Generazione dei Dati: Produce periodicamente misurazioni multivariate (es. temperatura, umidità) basate su distribuzioni realistiche per replicare l'ambiente fisico.

2.  Simulazione delle Anomalie: Introduce intenzionalmente outlier e dati mancanti per testare la robustezza e le capacità di filtraggio dei nodi Edge.

3.  Comunicazione Outgoing: Invia i dati grezzi all'Edge Hub designato utilizzando il protocollo MQTT.

4.  Qualità del Servizio: I messaggi di misurazione sono trasmessi con QoS at-most-once.

5.  Comunicazione di Controllo: Invia messaggi di registrazione contenenti i metadati del sensore con QoS exactly-once.

6.  Stato di Salute: Non trasmette heartbeat espliciti; la sua operatività è dedotta dalla continuità del flusso di misurazioni da parte dei nodi superiori.
*/
func main() {

	// Inizializza l'ambiente
	if err := environment.SetupEnvironment(); err != nil {
		println("Failed to setup environment:", err.Error())
		os.Exit(1)
	}

	// Inizializza il logger con il contesto
	logger.CreateLogger(logger.GetSensorAgentContext(environment.EdgeMacrozone, environment.EdgeZone, environment.SensorId))
	logger.PrintCurrentLevel()
	logger.Log.Info("Starting Sensor Agent...")
	logger.Log.Info("Sensor Location: ", environment.SensorLocation)
	logger.Log.Info("Sensor Type: ", environment.SensorType)
	logger.Log.Info("Sensor Reference: ", environment.SimulationSensorReference)

	// Registra il sensore all'edge hub
	comunication.SendRegistrationMessage()
	logger.Log.Info("Sensor registration message sent.")

	// Inizializza la comunicazione con il simulatore del sensore
	sensorChannelSource := make(chan types.SensorData, 100)
	go simulation.SimulateForever(sensorChannelSource)

	sensorChannelTarget := make(chan types.SensorData, 100)
	// Invia i dati al broker MQTT
	go comunication.PublishData(sensorChannelTarget)

	go func() {
		for data := range sensorChannelSource {
			health.UpdateLastValueTimestamp()
			// Invia i dati al canale di comunicazione
			select {
			case sensorChannelTarget <- data:
				health.UpdateLastValueTimestamp()
			default:
				logger.Log.Warn("MQTT channel is full, discarding data: ", data)
			}
		}
	}()

	// Abilita il canale di comunicazione per health check
	if environment.HealthzServer {
		logger.Log.Info("Enabling health check channel on port " + environment.HealthzServerPort)
		go func() {
			if err := health.StartHealthCheckServer(":" + environment.HealthzServerPort); err != nil {
				logger.Log.Error("Failed to enable health check channel: ", err.Error())
				os.Exit(1)
			}
		}()
	}

	utils.WaitForTerminationSignal()

	logger.Log.Info("Sensor agent is terminating")
	logger.Log.Info("Shutting down Sensor Agent...")

}
