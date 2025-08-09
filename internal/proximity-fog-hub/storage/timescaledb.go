package storage

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Definiamo una struttura per i risultati delle aggregazioni

type AggregatedStats struct {
	Timestamp string  `json:"timestamp"`
	Macrozone string  `json:"macrozone"`
	Type      string  `json:"type"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Avg       float64 `json:"avg"`
}

var DBPool *pgxpool.Pool

// InitDBConnection inizializza il pool di connessioni al database
func InitDatabaseConnection() error {
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		environment.PostgresUser, environment.PostgresPass, environment.PostgresHost, environment.PostgresPort, environment.PostgresDatabase,
	)
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("Unable to connect to database:  %w", err)
	}

	// Verifica la connessione
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	DBPool = pool
	logger.Log.Info("Connection to TimescaleDB for local cache successfully established.")
	return nil
}

// InsertSensorData inserisce un nuovo dato nella tabella della cache
func InsertSensorData(ctx context.Context, d types.SensorData) error {
	query := `
        INSERT INTO proximity_hub_measurements (time, building_id, floor_id, sensor_id, type, value)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := DBPool.Exec(ctx, query, d.Timestamp, d.EdgeMacrozone, d.EdgeZone, d.SensorID, d.Type, d.Data)
	return err
}

// GetStatsLastFiveMinutes interroga il DB per calcolare min, max e avg degli ultimi 5 minuti
func GetValueToSend(ctx context.Context) ([]AggregatedStats, error) {
	query := `
        SELECT 
            type,
            MIN(value) as min_val,
            MAX(value) as max_val,
            AVG(value) as avg_val
        FROM proximity_hub_measurements
        WHERE time >= NOW() - INTERVAL '5 minutes'
        GROUP BY type
    `
	rows, err := DBPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("aggregation query failed: %w", err)
	}
	defer rows.Close()

	var stats []AggregatedStats
	for rows.Next() {
		var s AggregatedStats
		if err := rows.Scan(&s.Type, &s.Min, &s.Max, &s.Avg); err != nil {
			logger.Log.Error("Error scanning statistics row, error:", err)
			continue
		}
		stats = append(stats, s)
	}

	return stats, nil
}
