package intermediate_fog_hub

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ProcessProximityFogHubData(dataChannel chan structure.SensorData) {

	// Connessione al DB
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		environment.PostgresSensorUser, environment.PostgresSensorPass, environment.PostgresSensorHost, environment.PostgresSensorPort, environment.PostgresSensorDatabase,
	)
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
	_, err := db.Exec(ctx, query, d.Timestamp, d.BuildingID, d.FloorID, d.SensorID, d.Type, d.Data)
	return err
}

func ProcessProximityFogHubConfiguration(msgChannel chan structure.ConfigurationMsg) {

	// Connessione al DB
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		environment.PostgresRegionUser, environment.PostgresRegionPass, environment.PostgresRegionHost, environment.PostgresRegionPort, environment.PostgresRegionDatabase,
	)
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	for msg := range msgChannel {
		logger.Log.Info("Received configuration message: ", msg)

		if msg.MsgType == "new_building" {
			logger.Log.Debug("Registration of new building: ", msg.BuildingID, ". Message type: ", msg.MsgType)
			if err := registerNewBuilding(ctx, pool, msg); err != nil {
				log.Fatalf("Registration failed: %v", err)
			}
		} else {
			logger.Log.Error("Invalid configuration message type: ", msg.MsgType)
		}

		logger.Log.Info("Configuration message processed: ", msg)

	}
}

func registerNewBuilding(ctx context.Context, db *pgxpool.Pool, d structure.ConfigurationMsg) error {
	query := `
        INSERT INTO buildings (name, location, registration_time, last_comunication)
        VALUES ($1, 
            ST_SetSRID(
                ST_MakePoint(
                    (random() * (15.0 - 6.0) + 6.0),   -- longitudine casuale tra 6 e 15
                    (random() * (47.0 - 36.0) + 36.0)  -- latitudine casuale tra 36 e 47
                ), 4326
            ),
            TO_TIMESTAMP($2),
            TO_TIMESTAMP($2)
        )
    `
	_, err := db.Exec(ctx, query, d.BuildingID, d.Timestamp)
	return err
}
