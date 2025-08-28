package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"

	"SensorContinuum/pkg/logger"
)

type PostgresDB struct {
	conn *pgx.Conn
}

var (
	cloudInstance *PostgresDB
	cloudOnce     sync.Once
	cloudInitErr  error

	regionInstances = make(map[string]*PostgresDB)
	regionOnce      = make(map[string]*sync.Once)
	regionInitErr   = make(map[string]error)

	sensorInstances = make(map[string]*PostgresDB)
	sensorOnce      = make(map[string]*sync.Once)
	sensorInitErr   = make(map[string]error)
)

// GetCloudPostgresDB Funzione per DB Cloud
func GetCloudPostgresDB(ctx context.Context) (*PostgresDB, error) {
	cloudOnce.Do(func() {
		dbURL := "postgres://admin:adminpass@metadata-db.cloud.sensorcontinuum.node:5433/sensorcontinuum"
		logger.Log.Info("Connecting to Cloud Postgres at ", dbURL)
		conn, err := pgx.Connect(ctx, dbURL)
		if err != nil {
			logger.Log.Error("Failed to connect to Cloud Postgres: ", err)
			cloudInitErr = err
			return
		}
		cloudInstance = &PostgresDB{conn: conn}
	})
	return cloudInstance, cloudInitErr
}

// GetRegionPostgresDB Funzione per DB Metadati Regione
func GetRegionPostgresDB(ctx context.Context, region string) (*PostgresDB, error) {
	if regionOnce[region] == nil {
		regionOnce[region] = &sync.Once{}
	}
	regionOnce[region].Do(func() {
		dbURL := fmt.Sprintf("postgres://admin:adminpass@metadata-db.%s.sensorcontinuum.node:5434/sensorcontinuum", region)
		logger.Log.Info("Connecting to Region Metadata Postgres at ", dbURL)
		conn, err := pgx.Connect(ctx, dbURL)
		if err != nil {
			logger.Log.Error("Failed to connect to Region Metadata Postgres: ", err)
			regionInitErr[region] = err
			return
		}
		regionInstances[region] = &PostgresDB{conn: conn}
	})
	return regionInstances[region], regionInitErr[region]
}

// Funzione per DB Misurazioni Regione
func GetSensorPostgresDB(ctx context.Context, region string) (*PostgresDB, error) {
	if sensorOnce[region] == nil {
		sensorOnce[region] = &sync.Once{}
	}
	sensorOnce[region].Do(func() {
		dbURL := fmt.Sprintf("postgres://admin:adminpass@mesurament-db.%s.sensorcontinuum.node:5432/sensorcontinuum", region)
		logger.Log.Info("Connecting to Region Sensor Postgres at ", dbURL)
		conn, err := pgx.Connect(ctx, dbURL)
		if err != nil {
			logger.Log.Error("Failed to connect to Region Sensor Postgres: ", err)
			sensorInitErr[region] = err
			return
		}
		sensorInstances[region] = &PostgresDB{conn: conn}
	})
	return sensorInstances[region], sensorInitErr[region]
}

func (db *PostgresDB) Close(ctx context.Context) error {
	if db.conn != nil {
		logger.Log.Info("Closing Postgres connection")
		return db.conn.Close(ctx)
	}
	return nil
}

func (db *PostgresDB) Conn() *pgx.Conn {
	return db.conn
}
