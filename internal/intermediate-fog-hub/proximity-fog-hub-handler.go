package intermediate_fog_hub

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/internal/intermediate-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ProcessRealTimeData(dataChannel chan types.SensorData) {

	err := storage.SetupSensorDbConnection()
	if err != nil {
		logger.Log.Error("Failed to connect to the sensor database: ", err)
		os.Exit(1)
	}
	defer storage.CloseSensorDbConnection()

	err = storage.SetupRegionDbConnection()
	if err != nil {
		logger.Log.Error("Failed to connect to the region database: ", err)
		os.Exit(1)
	}
	defer storage.CloseRegionDbConnection()

	var batch = types.NewSensorDataBatch()

	// Timer per la scrittura dei dati in batch
	timer := time.NewTimer(time.Second * time.Duration(environment.SensorDataBatchTimeout))

	for {
		select {
		case data := <-dataChannel:
			logger.Log.Info("Received real-time Data from: ", data.SensorID)

			if batch.Count() >= environment.SensorDataBatchSize {
				logger.Log.Info("Inserting batch data into the database")
				if err = storage.InsertSensorDataBatch(batch); err != nil {
					logger.Log.Error("Failed to insert sensor data batch: ", err)
				}
				logger.Log.Info("Updating last seen for batch sensors")
				if err = storage.UpdateLastSeenBatch(batch); err != nil {
					logger.Log.Error("Failed to update last seen for sensors: ", err)
				}
				logger.Log.Info("Clearing batch after processing")
				batch.Clear()
				timer.Reset(time.Second * 10)
				continue
			}

			logger.Log.Debug("Adding sensor data to batch: ", data.SensorID)
			batch.AddSensorData(data)

		case <-timer.C:
			if batch.Count() > 0 {
				logger.Log.Info("Inserting batch data into the database")
				if err = storage.InsertSensorDataBatch(batch); err != nil {
					logger.Log.Error("Failed to insert sensor data batch: ", err)
				}
				logger.Log.Info("Updating last seen for batch sensors")
				if err = storage.UpdateLastSeenBatch(batch); err != nil {
					logger.Log.Error("Failed to update last seen for sensors: ", err)
				}
				logger.Log.Info("Clearing batch after processing")
				batch.Clear()
			}
			timer.Reset(time.Second * 10)
		}
	}
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
