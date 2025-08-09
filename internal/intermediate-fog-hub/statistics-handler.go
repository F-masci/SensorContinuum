package intermediate_fog_hub

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProcessStatisticsData gestisce le statistiche aggregate e le salva.
func ProcessStatisticsData(statsChannel chan types.AggregatedStats) {
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		environment.PostgresSensorUser, environment.PostgresSensorPass, environment.PostgresSensorHost, environment.PostgresSensorPort, environment.PostgresSensorDatabase,
	)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Log.Error("Unable to connect to the database for statistics handler: ", err)
		return // Termina la goroutine se non può connettersi
	}
	defer pool.Close()

	for stats := range statsChannel {
		logger.Log.Info("Aggregated statistics received - type", stats.Type, " - avg", stats.Avg)
		if err := insertStatisticsData(ctx, pool, stats); err != nil {
			logger.Log.Error("Failed to insert statistics: ", err)
			// Non usiamo Fatalf per non far crashare l'intero servizio
		} else {
			logger.Log.Info("Statistics successfully inserted into the database")
		}
	}
}

func insertStatisticsData(ctx context.Context, db *pgxpool.Pool, s types.AggregatedStats) error {
	query := `
        INSERT INTO aggregated_statistics (time, macrozone_name, type, min_value, max_value, avg_value)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := db.Exec(ctx, query, s.Timestamp, s.Macrozone, s.Type, s.Min, s.Max, s.Avg)
	return err
}
