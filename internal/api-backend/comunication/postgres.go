package comunication

import (
	"context"
	"sync"

	"SensorContinuum/pkg/logger"
)

type PostgresDB struct {
	conn *pgx.Conn
}

var (
	instance *PostgresDB
	once     sync.Once
	initErr  error
)

const dbURL = "postgres://admin:adminpass@metadata-db.sensorcontinuum.node:5433/sensorcontinuum"

func GetPostgresDB(ctx context.Context) (*PostgresDB, error) {
	once.Do(func() {
		logger.Log.Info("Attempting to connect to Postgres at metadata-db.sensorcontinuum.node:5433")
		conn, err := pgx.Connect(ctx, dbURL)
		if err != nil {
			logger.Log.Error("Failed to connect to Postgres: ", err)
			initErr = err
			return
		}
		logger.Log.Info("Successfully connected to Postgres")
		instance = &PostgresDB{conn: conn}
	})
	return instance, initErr
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
