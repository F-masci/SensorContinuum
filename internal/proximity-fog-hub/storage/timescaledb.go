package storage

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DBPool *pgxpool.Pool

// InitDatabaseConnection inizializza il pool di connessioni al database
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
        INSERT INTO proximity_hub_measurements (time, macrozone_name, zone_name, sensor_id, type, value)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	t := time.Unix(d.Timestamp, 0).UTC()
	_, err := DBPool.Exec(ctx, query, t, d.EdgeMacrozone, d.EdgeZone, d.SensorID, d.Type, d.Data)
	return err
}

func GetZoneAggregatedData(ctx context.Context, start time.Time, end time.Time) ([]types.AggregatedStats, error) {
	query := `
        SELECT 
            type,
            zone_name,
            MIN(value) as min_val,
            MAX(value) as max_val,
            AVG(value) as avg_val,
            SUM(value) as avg_sum_val,
        	COUNT(value) as avg_count_val
        FROM proximity_hub_measurements
        WHERE time >= $1 AND time < $2 -- Usa i parametri di inizio e fine
        GROUP BY type, zone_name
    `
	rows, err := DBPool.Query(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("aggregation query failed: %w", err)
	}
	defer rows.Close()

	var stats []types.AggregatedStats
	for rows.Next() {
		var s types.AggregatedStats
		if err := rows.Scan(&s.Type, &s.Zone, &s.Min, &s.Max, &s.Avg, &s.Sum, &s.Count); err != nil {
			logger.Log.Error("Error scanning statistics row, error:", err)
			continue
		}
		s.Macrozone = environment.EdgeMacrozone
		stats = append(stats, s)
	}

	return stats, nil
}
