package intermediate_fog_hub

import (
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"

	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ProcessProximityFogHubData(dataChannel chan structure.SensorData) {

	// Connessione al DB
	dbURL := "postgres://admin:adminpass@localhost:5432/timeseriesdb?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	for data := range dataChannel {
		logger.Log.Info("Received Aggregated Data: ", data)

		if err := insertSensorData(ctx, pool, data); err != nil {
			log.Fatalf("Insert failed: %v", err)
		}

		logger.Log.Info("Data inserted successfully into the database: ", data)

	}
}

func insertSensorData(ctx context.Context, db *pgxpool.Pool, d structure.SensorData) error {
	query := `
        INSERT INTO sensor_measurements (time, building_id, floor_id, sensor_id, type, value)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := db.Exec(ctx, query, d.Timestamp, d.BuildingID, d.FloorID, d.SensorID, "temperature", d.Data)
	return err
}
