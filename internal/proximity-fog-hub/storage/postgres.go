package storage

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBPool è il pool di connessioni a TimescaleDB per la cache locale.
// La cache locale implementa il pattern Transactional Outbox
// per garantire che i dati aggregati vengano inviati in modo affidabile
// all' Intermediate Fog Hub tramite Kafka.
var DBPool *pgxpool.Pool

// aggregationLockConnection Connessione dedicata per il lock di aggregazione
// Usata per mantenere il lock attivo finché la connessione è aperta
var aggregationLockConnection *pgx.Conn

// dispatcherLockConnection Connessione dedicata per il lock del dispatcher
// Usata per mantenere il lock attivo finché la connessione è aperta
var dispatcherLockConnection *pgx.Conn

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

	// Connessione dedicata per il lock di aggregazione
	aggregationLockConnection, err = pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database for aggregation lock: %w", err)
	}

	// Verifica la connessione dedicata
	if err := aggregationLockConnection.Ping(ctx); err != nil {
		return fmt.Errorf("unable to connect to database for aggregation lock: %w", err)
	}

	// Connessione dedicata per il lock del dispatcher
	dispatcherLockConnection, err = pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database for dispatcher lock: %w", err)
	}

	// Verifica la connessione dedicata
	if err := dispatcherLockConnection.Ping(ctx); err != nil {
		return fmt.Errorf("unable to connect to database for dispatcher lock: %w", err)
	}

	logger.Log.Info("Connection to TimescaleDB for local cache successfully established.")
	return nil
}

// TryAcquireDispatcherLock prova ad acquisire il lock in Postgres.
// Restituisce true se il processo è leader, false altrimenti.
func TryAcquireDispatcherLock(ctx context.Context) (bool, error) {
	var gotLock bool
	err := dispatcherLockConnection.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", environment.DispatcherLockId).Scan(&gotLock)
	if err != nil {
		return false, fmt.Errorf("failed to acquire advisory lock: %w", err)
	}
	return gotLock, nil
}

// ReleaseDispatcherLock rilascia il lock (opzionale: si rilascia anche chiudendo la connessione)
func ReleaseDispatcherLock(ctx context.Context) error {
	_, err := dispatcherLockConnection.Exec(ctx, "SELECT pg_advisory_unlock($1)", environment.DispatcherLockId)
	return err
}

/* ----------- TRANSACTIONAL OUTBOX PATTERN ----------- */
/*			   		  DATI GREZZI 						*/
/* ---------------------------------------------------- */

// InsertSensorData inserisce un nuovo dato nella tabella della cache locale
func InsertSensorData(ctx context.Context, d types.SensorData) error {
	query := `
        INSERT INTO sensor_measurements_cache (time, macrozone_name, zone_name, sensor_id, type, value, status)
        VALUES ($1, $2, $3, $4, $5, $6, 'pending')
        ON CONFLICT (time, macrozone_name, zone_name, sensor_id, type) DO NOTHING
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
	defer func() {
		if err != nil && tx != nil {
			if rerr := tx.Rollback(ctx); rerr != nil && !errors.Is(rerr, pgx.ErrTxClosed) {
				logger.Log.Error("failed to rollback transaction: ", rerr)
				os.Exit(1)
			}
		}
	}()

	query := `
		UPDATE sensor_measurements_cache
		SET status = $1
		WHERE time = $2
		  	AND macrozone_name = $3
		  	AND zone_name = $4
			AND sensor_id = $5
			AND type = $6
	`
	for _, d := range data {
		t := time.Unix(d.Timestamp, 0).UTC()
		_, err := tx.Exec(ctx, query, newStatus, t, d.EdgeMacrozone, d.EdgeZone, d.SensorID, d.Type)
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
		ON CONFLICT (time, zone_name, type) DO NOTHING;
	`
	t := time.Unix(stats.Timestamp, 0).UTC()
	_, err := DBPool.Exec(ctx, query, t, stats.Zone, stats.Type, stats.Min, stats.Max, stats.Avg, stats.Sum, stats.Count)
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
	defer func() {
		if err != nil && tx != nil {
			if rerr := tx.Rollback(ctx); rerr != nil && !errors.Is(rerr, pgx.ErrTxClosed) {
				logger.Log.Error("failed to rollback transaction: ", rerr)
				os.Exit(1)
			}
		}
	}()

	query := `
		UPDATE aggregated_stats_cache
		SET status = $1
		WHERE time = $2 
		  AND zone_name = $3
		  AND type = $4
	`
	for _, d := range data {
		t := time.Unix(d.Timestamp, 0).UTC()
		_, err := tx.Exec(ctx, query, newStatus, t, d.Zone, d.Type)
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

// TryAcquireAggregationLock prova ad acquisire il lock in Postgres.
// Restituisce true se il processo è leader, false altrimenti.
func TryAcquireAggregationLock(ctx context.Context) (bool, error) {
	var gotLock bool
	err := aggregationLockConnection.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", environment.AggregationLockId).Scan(&gotLock)
	if err != nil {
		return false, fmt.Errorf("failed to acquire advisory lock: %w", err)
	}
	return gotLock, nil
}

// ReleaseAggregationLock rilascia il lock (opzionale: si rilascia anche chiudendo la connessione)
func ReleaseAggregationLock(ctx context.Context) error {
	_, err := aggregationLockConnection.Exec(ctx, "SELECT pg_advisory_unlock($1)", environment.AggregationLockId)
	return err
}

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
		WHERE zone_name = ''
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
