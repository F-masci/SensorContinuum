package storage

import (
	"SensorContinuum/pkg/types"
	"context"
	"fmt"
	"sync"
	"time"

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
		dbURL := "postgres://admin:adminpass@metadata-db.cloud.sensorcontinuum.local:5433/sensorcontinuum"
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
		dbURL := fmt.Sprintf("postgres://admin:adminpass@metadata-db.%s.sensorcontinuum.local:5434/sensorcontinuum", region)
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
		dbURL := fmt.Sprintf("postgres://admin:adminpass@mesurament-db.%s.sensorcontinuum.local:5432/sensorcontinuum", region)
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

/* ----- Funzioni per la gestione delle viste materializzate ----- */

// computeTruncatedTimestamp calcola il timestamp troncato in base alla vista e restituisce la granularità
func computeTruncatedTimestamp(viewName string, ts time.Time) (time.Time, string, error) {
	ts = ts.UTC() // forzare UTC

	switch viewName {
	case "macrozone_daily_agg", "region_daily_agg":
		return ts.Truncate(24 * time.Hour), "day", nil
	case "macrozone_weekly_agg", "region_weekly_agg":
		year, week := ts.ISOWeek()
		firstDay := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		offsetDays := (week - 1) * 7
		return firstDay.AddDate(0, 0, offsetDays), "week", nil
	case "macrozone_monthly_agg", "region_monthly_agg":
		return time.Date(ts.Year(), ts.Month(), 1, 0, 0, 0, 0, time.UTC), "month", nil
	case "macrozone_yearly_agg", "region_yearly_agg":
		return time.Date(ts.Year(), 1, 1, 0, 0, 0, 0, time.UTC), "year", nil
	default:
		return ts, "", fmt.Errorf("unknown view: %s", viewName)
	}
}

// recordExists controlla se il record per il timestamp esiste nella view
func recordExists(ctx context.Context, db *PostgresDB, viewName string, colName string, ts time.Time) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s = $1", viewName, colName)
	var count int
	if err := db.conn.QueryRow(ctx, query, ts).Scan(&count); err != nil {
		return false, fmt.Errorf("query error for %s: %w", viewName, err)
	}
	return count > 0, nil
}

// getBucketRange calcola l'inizio e la fine del bucket che contiene ts
func getBucketRange(viewName string, ts time.Time) (time.Time, time.Time, error) {
	ts = ts.UTC() // lavoriamo sempre in UTC

	switch viewName {
	// Macrozone daily/region daily
	case "macrozone_daily_agg", "region_daily_agg":
		start := time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 0, 1)
		return start, end, nil

	// Macrozone weekly/region weekly
	case "macrozone_weekly_agg", "region_weekly_agg":
		// Consideriamo settimana che inizia lunedì
		weekday := int(ts.Weekday())
		if weekday == 0 {
			weekday = 7 // domenica = 7
		}
		start := time.Date(ts.Year(), ts.Month(), ts.Day()-weekday+1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 0, 7)
		return start, end, nil

	// Macrozone monthly/region monthly
	case "macrozone_monthly_agg", "region_monthly_agg":
		start := time.Date(ts.Year(), ts.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)
		return start, end, nil

	// Macrozone yearly/region yearly
	case "macrozone_yearly_agg", "region_yearly_agg":
		start := time.Date(ts.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(1, 0, 0)
		return start, end, nil
	}

	return time.Time{}, time.Time{}, fmt.Errorf("unknown view name: %s", viewName)
}

// refreshAggregate calcola la finestra corretta e chiama la funzione di refresh
func refreshAggregate(ctx context.Context, db *PostgresDB, viewName string, ts time.Time) error {
	start, end, err := getBucketRange(viewName, ts)
	if err != nil {
		return err
	}

	// SQL: passiamo direttamente start e end come intervalli temporali
	query := fmt.Sprintf(
		"CALL refresh_continuous_aggregate('%s', '%s', '%s', true);",
		viewName, start.Format(time.RFC3339), end.Format(time.RFC3339),
	)

	if _, err := db.conn.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to refresh continuous aggregate for %s: %w", viewName, err)
	}

	return nil
}

// CheckAndRefreshAggregateView verifica se il record esiste in una vista materializzata e, se no, richiama refresh_continuous_aggregate
func CheckAndRefreshAggregateView(ctx context.Context, db *PostgresDB, viewName string, timestamp time.Time) error {
	truncatedTs, colName, err := computeTruncatedTimestamp(viewName, timestamp)
	if err != nil {
		return err
	}

	exists, err := recordExists(ctx, db, viewName, colName, truncatedTs)
	if err != nil {
		return err
	}

	if !exists {
		if err := refreshAggregate(ctx, db, viewName, truncatedTs); err != nil {
			return err
		}
	}

	return nil
}

/* ----- Salvataggio batch delle analisi ----- */

// SaveVariationResults salva in batch i risultati di tipo VariationResult
func SaveVariationResults(ctx context.Context, db *PostgresDB, results []types.VariationResult) error {
	if len(results) == 0 {
		return nil // niente da salvare
	}

	// Prepariamo i valori per l'INSERT
	rows := make([][]interface{}, 0, len(results))
	for _, r := range results {
		// Convertiamo timestamp UNIX -> UTC time
		ts := time.Unix(r.Timestamp, 0).UTC()
		rows = append(rows, []interface{}{
			ts.Truncate(24 * time.Hour), r.Macrozone, r.Type, r.Current, r.Previous, r.DeltaPerc,
		})
	}

	// Query parametrica in formato COPY-like
	_, err := db.conn.CopyFrom(
		ctx,
		pgx.Identifier{"macrozone_yearly_variation"},
		[]string{"time", "macrozone", "type", "current", "previous", "delta_perc"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("failed to save variation results: %w", err)
	}

	return nil
}

// SaveAnomaliesResults salva in batch i risultati di tipo MacrozoneAnomaly
func SaveAnomaliesResults(ctx context.Context, db *PostgresDB, results []types.MacrozoneAnomaly) error {
	if len(results) == 0 {
		return nil // niente da salvare
	}

	// Prepariamo i valori per l'INSERT
	rows := make([][]interface{}, 0, len(results))
	for _, r := range results {
		// Convertiamo timestamp UNIX -> UTC time
		ts := time.Unix(r.Timestamp, 0).UTC()
		rows = append(rows, []interface{}{
			ts.Truncate(24 * time.Hour), r.MacrozoneName, r.Type, r.Variation.Current,
			r.Variation.Previous, r.Variation.DeltaPerc, r.NeighborMean, r.NeighborStdDev,
			r.AbsError, r.ZScore,
		})
	}

	// Query parametrica in formato COPY-like
	_, err := db.conn.CopyFrom(
		ctx,
		pgx.Identifier{"macrozone_yearly_anomalies"},
		[]string{"time", "macrozone", "type", "current", "previous", "delta_perc",
			"neighbor_mean", "neighbor_std_dev", "abs_error", "z_score"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("failed to save anomalies results: %w", err)
	}

	return nil
}

// SaveTrendSimilarityResults salva in batch i risultati di tipo TrendSimilarityResult
func SaveTrendSimilarityResults(ctx context.Context, db *PostgresDB, results []types.TrendSimilarityResult, date time.Time) error {
	if len(results) == 0 {
		return nil // niente da salvare
	}

	// Prepariamo i valori per l'INSERT
	rows := make([][]interface{}, 0, len(results))
	for _, r := range results {
		// timestamp di riferimento (giorno passato alla funzione di calcolo)
		ts := date.UTC().Truncate(24 * time.Hour)
		rows = append(rows, []interface{}{
			ts, r.MacrozoneName, r.Type, r.Correlation, r.SlopeMacro, r.SlopeRegion, r.Divergence,
		})
	}

	// Inserimento bulk COPY
	_, err := db.conn.CopyFrom(
		ctx,
		pgx.Identifier{"macrozone_trends_similarity"},
		[]string{"time", "macrozone", "type", "correlation", "slope_macro", "slope_region", "divergence"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("failed to save trend similarity results: %w", err)
	}

	return nil
}
