package storage

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DBPool *pgxpool.Pool

// OutboxMessage rappresenta un record nella tabella aggregated_stats_outbox
type OutboxMessage struct {
	ID      uuid.UUID
	Payload types.AggregatedStats
	Status  string
}

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
        INSERT INTO sensor_measurements_cache (time, macrozone_name, zone_name, sensor_id, type, value)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	t := time.Unix(d.Timestamp, 0).UTC()
	_, err := DBPool.Exec(ctx, query, t, d.EdgeMacrozone, d.EdgeZone, d.SensorID, d.Type, d.Data)
	return err
}

// InsertAggregatedStatsOutbox inserisce un record di statistiche aggregate nella tabella outbox.
// Questa è la prima fase del pattern Transactional Outbox.
func InsertAggregatedStatsOutbox(ctx context.Context, stats types.AggregatedStats) error {
	// Serializziamo il payload in JSON
	payloadBytes, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal aggregated stats payload: %w", err)
	}

	query := ` 
        INSERT INTO aggregated_stats_outbox (payload)
        VALUES ($1)
    `
	_, err = DBPool.Exec(ctx, query, payloadBytes)
	if err != nil {
		return fmt.Errorf("failed to insert into outbox table: %w", err)
	}
	return nil
}

// GetPendingOutboxMessages recupera un batch di messaggi in stato 'pending' dalla tabella outbox.
// Utilizza SELECT ... FOR UPDATE SKIP LOCKED per garantire che worker concorrenti non prelevino gli stessi messaggi.
func GetPendingOutboxMessages(ctx context.Context, limit int) ([]OutboxMessage, error) {
	query := `
        SELECT id, payload
        FROM aggregated_stats_outbox
        WHERE status = 'pending'
        ORDER BY created_at
        LIMIT $1
        FOR UPDATE SKIP LOCKED
    `
	rows, err := DBPool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending outbox messages: %w", err)
	}
	defer rows.Close()

	var messages []OutboxMessage
	for rows.Next() {
		var msg OutboxMessage
		var payloadBytes []byte
		if err := rows.Scan(&msg.ID, &payloadBytes); err != nil {
			logger.Log.Error("Error scanning outbox message row, error:", err)
			continue
		}

		// Deserializziamo il payload JSON nella nostra struct
		if err := json.Unmarshal(payloadBytes, &msg.Payload); err != nil {
			logger.Log.Error("Error unmarshalling outbox payload, skipping message ID ", msg.ID, ", error:", err)
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// UpdateOutboxMessageStatus aggiorna lo stato di un messaggio nella tabella outbox.
// Solitamente viene chiamato dopo che il messaggio è stato inviato con successo a Kafka.
func UpdateOutboxMessageStatus(ctx context.Context, id uuid.UUID, newStatus string) error {
	query := `
        UPDATE aggregated_stats_outbox
        SET status = $1, updated_at = NOW()
        WHERE id = $2
    `
	_, err := DBPool.Exec(ctx, query, newStatus, id)
	if err != nil {
		return fmt.Errorf("failed to update outbox message status for ID %s: %w", id, err)
	}
	return nil
}

// DeleteSentOutboxMessages elimina i messaggi dalla tabella outbox che sono già stati inviati
// e sono più vecchi di una certa durata. Questo previene che la tabella cresca indefinitamente.
func DeleteSentOutboxMessages(ctx context.Context, olderThan time.Duration) (int64, error) {
	// Calcoliamo il timestamp limite.
	// olderThan specifica quanto vecchio deve essere un messaggio prima di dover essere cancellato
	// Esempio: se olderThan è 1 ora, cancelliamo i messaggi il cui ultimo aggiornamento (updated_at)
	// è avvenuto più di un'ora fa.
	timestampLimit := time.Now().UTC().Add(-olderThan)

	query := `
        DELETE FROM aggregated_stats_outbox
        WHERE status = 'sent' AND updated_at < $1
    `
	commandTag, err := DBPool.Exec(ctx, query, timestampLimit)
	if err != nil {
		return 0, fmt.Errorf("failed to delete sent outbox messages: %w", err)
	}

	// commandTag.RowsAffected() restituisce il numero di righe eliminate.
	return commandTag.RowsAffected(), nil
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
