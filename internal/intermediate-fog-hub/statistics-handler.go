package intermediate_fog_hub

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProcessStatisticsData gestisce le statistiche aggregate e le salva.
func ProcessStatisticsData(statsChannel chan structure.AggregatedStats) {
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		environment.PostgresUser, environment.PostgresPass, environment.PostgresHost, environment.PostgresPort, environment.PostgresDatabase,
	)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Log.Error("Impossibile connettersi al database per il gestore di statistiche", "error", err)
		return // Termina la goroutine se non può connettersi
	}
	defer pool.Close()

	for stats := range statsChannel {
		logger.Log.Info("Statistiche aggregate ricevute", "type", stats.Type, "avg", stats.Avg)
		if err := insertStatisticsData(ctx, pool, stats); err != nil {
			logger.Log.Error("Inserimento statistiche fallito", "error", err)
			// Non usiamo Fatalf per non far crashare l'intero servizio
		} else {
			logger.Log.Info("Statistiche inserite con successo nel database")
		}
	}
}

func insertStatisticsData(ctx context.Context, db *pgxpool.Pool, s structure.AggregatedStats) error {
	query := `
        INSERT INTO aggregated_statistics (time, building_id, type, min_value, max_value, avg_value)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := db.Exec(ctx, query, s.Timestamp, s.BuildingID, s.Type, s.Min, s.Max, s.Avg)
	return err
}
