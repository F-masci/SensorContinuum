package storage

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBPool è il pool di connessioni a TimescaleDB per la cache locale.
// La cache locale implementa il pattern Transactional Outbox
// per garantire che i dati aggregati vengano inviati in modo affidabile
// all' Intermediate Fog Hub tramite Kafka.
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

/* ----------- TRANSACTIONAL OUTBOX PATTERN ----------- */
/*			   		  DATI GREZZI 						*/
/* ---------------------------------------------------- */

// InsertSensorData inserisce un nuovo dato nella tabella della cache locale
func InsertSensorData(ctx context.Context, d types.SensorData) error {
	query := `
        INSERT INTO sensor_measurements_cache (time, macrozone_name, zone_name, sensor_id, type, value, status)
        VALUES ($1, $2, $3, $4, $5, $6, 'pending')
    `
	t := time.Unix(d.Timestamp, 0).UTC()
	_, err := DBPool.Exec(ctx, query, t, d.EdgeMacrozone, d.EdgeZone, d.SensorID, d.Type, d.Data)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Se l'errore è per violazione di chiave unica (duplicate), lo ignoriamo
			logger.Log.Warn("Duplicate sensor data entry for ", d.SensorID, "_", t.Format(time.RFC3339), ", skipping insert.")
			return nil
		}
		return fmt.Errorf("failed to insert aggregated stats: %w", err)
	}
	return nil
}

// GetPendingSensorData recupera un batch di messaggi in stato 'pending' dalla tabella outbox.
// Garantisce che worker concorrenti non prelevino gli stessi messaggi.
func GetPendingSensorData(ctx context.Context, limit int) ([]types.SensorData, error) {

	// Utilizza SELECT ... FOR UPDATE SKIP LOCKED per garantire che worker concorrenti non prelevino gli stessi messaggi.
	query := `
		SELECT 
			time, macrozone_name, zone_name, sensor_id, type, value
		FROM sensor_measurements_cache
		WHERE status = 'pending'
		ORDER BY time
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	rows, err := DBPool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending outbox messages: %w", err)
	}
	defer rows.Close()

	var messages []types.SensorData
	for rows.Next() {
		var msg types.SensorData
		var t time.Time
		if err := rows.Scan(&t, &msg.EdgeMacrozone, &msg.EdgeZone, &msg.SensorID, &msg.Type, &msg.Data); err != nil {
			logger.Log.Error("Error scanning outbox message row, error:", err)
			continue
		}
		msg.Timestamp = t.UTC().Unix()
		messages = append(messages, msg)
	}

	return messages, nil
}

// UpdateSensorData aggiorna lo stato di un messaggio nella tabella outbox.
// Solitamente viene chiamato dopo che il messaggio è stato inviato con successo a Kafka.
func UpdateSensorData(ctx context.Context, data []types.SensorData, newStatus string) error {
	tx, err := DBPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE sensor_measurements_cache
		SET status = $1
		WHERE time = $2 AND sensor_id IS NOT DISTINCT FROM $3
	`
	for _, d := range data {
		t := time.Unix(d.Timestamp, 0).UTC()
		_, err := tx.Exec(ctx, query, newStatus, t, d.SensorID)
		if err != nil {
			return fmt.Errorf("failed to update outbox message status for %s_%s: %w", d.SensorID, t.Format(time.RFC3339), err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// DeleteSensorData elimina i messaggi dalla tabella outbox che sono già stati inviati
// e sono più vecchi di una certa durata. Questo previene che la tabella cresca indefinitamente.
func DeleteSensorData(ctx context.Context, olderThan time.Duration) (int64, error) {
	// Calcoliamo il timestamp limite.
	// olderThan specifica quanto vecchio deve essere un messaggio prima di dover essere cancellato
	// Esempio: se olderThan è 1 ora, cancelliamo i messaggi il cui ultimo aggiornamento (updated_at)
	// è avvenuto più di un'ora fa.
	timestampLimit := time.Now().UTC().Add(-olderThan)
	query := `
		DELETE FROM sensor_measurements_cache
		WHERE status = 'sent' AND time < $1
	`
	commandTag, err := DBPool.Exec(ctx, query, timestampLimit)
	if err != nil {
		return 0, fmt.Errorf("failed to delete sent outbox messages: %w", err)
	}

	// restituisce il numero di righe eliminate.
	return commandTag.RowsAffected(), nil
}

/* ----------- TRANSACTIONAL OUTBOX PATTERN ----------- */
/*			   		  DATI AGGREGATI 					*/
/* ---------------------------------------------------- */

// InsertAggregatedStats inserisce un record di statistiche aggregate nella cache locale
func InsertAggregatedStats(ctx context.Context, stats types.AggregatedStats) error {
	query := `
		INSERT INTO aggregated_stats_cache (time, zone_name, type, min_value, max_value, avg_value, avg_sum, avg_count, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending')
	`
	t := time.Unix(stats.Timestamp, 0).UTC()
	var z any
	if stats.Zone == "" {
		z = nil // imposta NULL
	} else {
		z = stats.Zone
	}
	_, err := DBPool.Exec(ctx, query, t, z, stats.Type, stats.Min, stats.Max, stats.Avg, stats.Sum, stats.Count)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Se l'errore è per violazione di chiave unica (duplicate), lo ignoriamo
			logger.Log.Warn("Duplicate aggregated stats entry for ", stats.Zone, "_", t.Format(time.RFC3339), ", skipping insert.")
			return nil
		}
		return fmt.Errorf("failed to insert aggregated stats: %w", err)
	}
	return nil
}

// GetPendingAggregatedStats recupera un batch di messaggi in stato 'pending' dalla tabella outbox.
// Garantisce che worker concorrenti non prelevino gli stessi messaggi.
func GetPendingAggregatedStats(ctx context.Context, limit int) ([]types.AggregatedStats, error) {

	// Utilizza SELECT ... FOR UPDATE SKIP LOCKED per garantire che worker concorrenti non prelevino gli stessi messaggi.
	query := `
        SELECT 
            time, zone_name, type, min_value, max_value, avg_value, avg_sum, avg_count
        FROM aggregated_stats_cache
        WHERE status = 'pending'
        ORDER BY time
        LIMIT $1
        FOR UPDATE SKIP LOCKED
    `
	rows, err := DBPool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending outbox messages: %w", err)
	}
	defer rows.Close()

	var messages []types.AggregatedStats
	for rows.Next() {
		var msg types.AggregatedStats
		var t time.Time
		var z any
		// zone_name può essere NULL, quindi usiamo 'any' e gestiamo di conseguenza durante la scansione.
		if err := rows.Scan(&t, &z, &msg.Type, &msg.Min, &msg.Max, &msg.Avg, &msg.Sum, &msg.Count); err != nil {
			logger.Log.Error("Error scanning outbox message row, error:", err)
			continue
		}
		if z == nil {
			msg.Zone = ""
		} else {
			msg.Zone = z.(string)
		}
		msg.Timestamp = t.UTC().Unix()
		msg.Macrozone = environment.EdgeMacrozone
		messages = append(messages, msg)
	}

	return messages, nil
}

// UpdateAggregatedStats aggiorna lo stato di un messaggio nella tabella outbox.
// Solitamente viene chiamato dopo che il messaggio è stato inviato con successo a Kafka.
func UpdateAggregatedStats(ctx context.Context, data []types.AggregatedStats, newStatus string) error {
	tx, err := DBPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE aggregated_stats_cache
		SET status = $1
		WHERE time = $2 AND zone_name IS NOT DISTINCT FROM $3
	`
	for _, d := range data {
		t := time.Unix(d.Timestamp, 0).UTC()
		var z any
		if d.Zone == "" {
			z = nil
		} else {
			z = d.Zone
		}
		_, err := tx.Exec(ctx, query, newStatus, t, z)
		if err != nil {
			return fmt.Errorf("failed to update outbox message status for %s_%s: %w", d.Zone, t.Format(time.RFC3339), err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// DeleteAggregatedStats elimina i messaggi dalla tabella outbox che sono già stati inviati
// e sono più vecchi di una certa durata. Questo previene che la tabella cresca indefinitamente.
func DeleteAggregatedStats(ctx context.Context, olderThan time.Duration) (int64, error) {
	// Calcoliamo il timestamp limite.
	// olderThan specifica quanto vecchio deve essere un messaggio prima di dover essere cancellato
	// Esempio: se olderThan è 1 ora, cancelliamo i messaggi il cui ultimo aggiornamento (updated_at)
	// è avvenuto più di un'ora fa.
	timestampLimit := time.Now().UTC().Add(-olderThan)

	query := `
        DELETE FROM aggregated_stats_cache
        WHERE status = 'sent' AND time < $1
    `
	commandTag, err := DBPool.Exec(ctx, query, timestampLimit)
	if err != nil {
		return 0, fmt.Errorf("failed to delete sent outbox messages: %w", err)
	}

	// restituisce il numero di righe eliminate.
	return commandTag.RowsAffected(), nil
}

/* ----------- DATI AGGREGATI ----------- */

// GetZoneAggregatedData calcola le statistiche aggregate (min, max, avg, sum, count)
// per ogni tipo di sensore e per ogni zona, nell'intervallo di tempo specificato.
// Restituisce una slice di AggregatedStats, una per ogni combinazione di tipo e zona.
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
        FROM sensor_measurements_cache
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

// GetLastMacrozoneAggregatedData ritorna le ultime statistiche aggregate a livello di macrozona
func GetLastMacrozoneAggregatedData(ctx context.Context) (types.AggregatedStats, error) {
	query := `
		SELECT 
		    time,
			type,
			min_value,
			max_value,
			avg_value,
			avg_sum,
			avg_count
		FROM aggregated_stats_cache
		WHERE zone_name IS NULL
		ORDER BY time DESC
		LIMIT 1
	`
	rows, err := DBPool.Query(ctx, query)
	if err != nil {
		return types.AggregatedStats{}, fmt.Errorf("query for last macrozone aggregated data failed: %w", err)
	}
	defer rows.Close()

	var aggregatedStats types.AggregatedStats
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t, &aggregatedStats.Type, &aggregatedStats.Min, &aggregatedStats.Max, &aggregatedStats.Avg, &aggregatedStats.Sum, &aggregatedStats.Count); err != nil {
			logger.Log.Error("Error scanning last macrozone statistics row, error:", err)
			continue
		}
		aggregatedStats.Timestamp = t.UTC().Unix()
		aggregatedStats.Macrozone = environment.EdgeMacrozone
	}

	return aggregatedStats, nil
}
