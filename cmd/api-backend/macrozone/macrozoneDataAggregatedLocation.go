package main

import (
	"SensorContinuum/internal/api-backend/macrozone"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	ctx := context.Background()
	aggregatedStats, err := macrozone.GetAggregatedSensorDataByLocation(ctx, 41.8965, 12.4940, 1500)
	if err != nil {
		log.Fatal(err)
	}

	// Trasforma in JSON indentato
	dataJSON, err := json.MarshalIndent(aggregatedStats, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(dataJSON))
}
