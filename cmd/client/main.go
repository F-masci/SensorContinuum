package main

import (
	"SensorContinuum/internal/client/cli"
	"SensorContinuum/internal/client/environment"
	"SensorContinuum/pkg/logger"
)

func main() {

	// Inizializza il logger per l'applicazione
	// Questo logger è configurato per registrare i messaggi di debug e di errore
	logger.CreateLogger(logger.Context{
		"service": "client",
		"module":  "main",
	})
	logger.SetLoggerLevel(logger.ErrorLevel)

	// Inizializza l'ambiente e le configurazioni necessarie
	err := environment.SetupEnvironment()
	if err != nil {
		// Gestione dell'errore di configurazione
		// Se non riesce a caricare le variabili d'ambiente, esce con un errore
		panic("Errore nella configurazione dell'ambiente: " + err.Error())
	}

	// Avvia il menù principale dell'applicazione
	// Questo menù è stateless e non mantiene lo stato tra le chiamate
	// Le operazioni come la gestione delle regioni saranno gestite in modo stateless
	// all'interno del menù stesso, senza mantenere uno stato persistente
	// per ogni operazione.
	cli.MainMenu()
}
