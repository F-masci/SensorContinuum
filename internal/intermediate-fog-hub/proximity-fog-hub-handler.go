package intermediate_fog_hub

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ProcessProximityFogHubData(dataChannel chan types.SensorData) {

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

func insertSensorData(ctx context.Context, db *pgxpool.Pool, d types.SensorData) error {
	query := `
        INSERT INTO sensor_measurements (time, building_id, floor_id, sensor_id, type, value)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := db.Exec(ctx, query, d.Timestamp, d.EdgeMacrozone, d.EdgeZone, d.SensorID, d.Type, d.Data)
	return err
}

func ProcessProximityFogHubConfiguration(msgChannel chan types.ConfigurationMsg) {

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

		if msg.MsgType == types.NewProximityMsgType {
			logger.Log.Debug("Registration of new macrozone hub: ", msg.EdgeMacrozone, ". Hub ID: ", msg.HubID)
			if err := registerMacrozoneHub(ctx, pool, msg); err != nil {
				log.Fatalf("Registration failed: %v", err)
			}
		} else if msg.MsgType == types.NewEdgeMsgType {
			logger.Log.Debug("Registration of new zone hub: ", msg.EdgeZone, ". Hub ID: ", msg.HubID)
			if err := registerZoneHub(ctx, pool, msg); err != nil {
				log.Fatalf("Registration failed: %v", err)
			}
		} else if msg.MsgType == types.NewSensorMsgType {
			logger.Log.Debug("Registration of new sensor: ", msg.SensorID, ". Zone Hub ID: ", msg.HubID)
			if err := registerSensor(ctx, pool, msg); err != nil {
				log.Fatalf("Registration failed: %v", err)
			}
		} else {
			logger.Log.Error("Invalid configuration message types: ", msg.MsgType)
		}

		logger.Log.Info("Configuration message processed: ", msg)

	}
}

// Registra o aggiorna un hub di macrozona (proximity fog hub)
func registerMacrozoneHub(ctx context.Context, db *pgxpool.Pool, msg types.ConfigurationMsg) error {
	timestamp := time.Unix(msg.Timestamp, 0)
	query := `
		INSERT INTO macrozone_hubs (id, macrozone_name, service, registration_time, last_seen)
		VALUES ($1, $2, $3, $4, $4)
		ON CONFLICT (id, macrozone_name) DO UPDATE SET last_seen = EXCLUDED.last_seen
	`
	_, err := db.Exec(ctx, query, msg.HubID, msg.EdgeMacrozone, msg.Service, timestamp)
	return err
}

// Registra o aggiorna un hub di zona (edge hub)
func registerZoneHub(ctx context.Context, db *pgxpool.Pool, msg types.ConfigurationMsg) error {
	timestamp := time.Unix(msg.Timestamp, 0)
	query := `
		INSERT INTO zone_hubs (id, macrozone_name, zone_name, service, registration_time, last_seen)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (id, macrozone_name, zone_name) DO UPDATE SET last_seen = EXCLUDED.last_seen
	`
	_, err := db.Exec(ctx, query, msg.HubID, msg.EdgeMacrozone, msg.EdgeZone, msg.Service, timestamp)
	return err
}

// Registra o aggiorna un sensore associato a un edge hub
func registerSensor(ctx context.Context, db *pgxpool.Pool, msg types.ConfigurationMsg) error {
	timestamp := time.Unix(msg.Timestamp, 0)
	query := `
        INSERT INTO sensors (id, macrozone_name, zone_name, type, reference, registration_time, last_seen)
        VALUES ($1, $2, $3, $4, $5, $6, $6)
        ON CONFLICT (id, macrozone_name, zone_name) DO UPDATE SET last_seen = EXCLUDED.last_seen
    `
	_, err := db.Exec(ctx, query, msg.SensorID, msg.EdgeMacrozone, msg.EdgeZone, msg.SensorType, msg.SensorReference, timestamp)
	return err
}

func Register() error {

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

	query := `
	INSERT INTO region_hubs (id, service, registration_time, last_seen)
	VALUES ($1, $2, NOW(), NOW())
	ON CONFLICT (id) DO UPDATE SET last_seen = EXCLUDED.last_seen
`
	_, err = pool.Exec(ctx, query, environment.HubID, types.IntrermediateHubService)
	return err
}
