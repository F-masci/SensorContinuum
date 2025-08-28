package main

import (
	macrozoneAPI "SensorContinuum/internal/api-backend/macrozone"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	ctx := context.Background()

	// 1. Ottieni tutte le macrozone della regione
	macrozones, err := macrozoneAPI.GetMacrozonesList(ctx, "region-001")
	if err != nil {
		log.Fatal(err)
	}
	if len(macrozones) == 0 {
		log.Fatal("Nessuna macrozona trovata per la regione specificata")
	}

	macrozoneVariations, err := macrozoneAPI.CalculateAnomaliesPerMacrozones(ctx, macrozones, 100_000_000)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Salvare i dati

	// Trasforma in JSON indentato
	dataJSON, err := json.MarshalIndent(macrozoneVariations, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(dataJSON))
}
