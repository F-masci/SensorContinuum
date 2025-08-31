package storage

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// postgresDb incapsula la connessione al database PostgreSQL
type postgresDb struct {
	Db   *pgxpool.Pool
	Lock *pgx.Conn
	Ctx  context.Context
	Url  string
}

// Connect stabilisce la connessione al database PostgreSQL
func (p *postgresDb) Connect() error {

	// Se la connessione è già stabilita, non fare nulla
	if p.Db != nil {
		logger.Log.Debug("Database connection already established")
		return nil
	}

	logger.Log.Info("Connecting to the database at ", p.Url)
	var err error
	p.Db, err = pgxpool.New(p.Ctx, p.Url)
	if err != nil {
		logger.Log.Error("Unable to connect to the database: ", err)
		os.Exit(1)
	}

	// Verifica la connessione
	err = p.Db.Ping(p.Ctx)
	if err != nil {
		logger.Log.Error("Unable to ping the database: ", err)
		os.Exit(1)
	}

	// Connessione per il lock
	p.Lock, err = pgx.Connect(p.Ctx, p.Url)
	if err != nil {
		logger.Log.Error("Unable to connect to the database for locking: ", err)
		os.Exit(1)
	}

	// Verifica la connessione per il lock
	err = p.Lock.Ping(p.Ctx)
	if err != nil {
		logger.Log.Error("Unable to ping the database for locking: ", err)
		os.Exit(1)
	}

	logger.Log.Info("Connected to the database successfully")
	return nil
}

// TryBecomeLeader prova ad acquisire il lock in Postgres.
// Restituisce true se il processo è leader, false altrimenti.
func (p *postgresDb) TryBecomeLeader(ctx context.Context) (bool, error) {
	var gotLock bool
	err := p.Lock.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", environment.AggregationLockId).Scan(&gotLock)
	if err != nil {
		return false, fmt.Errorf("failed to acquire advisory lock: %w", err)
	}
	return gotLock, nil
}

// ReleaseLeadership rilascia il lock (opzionale: si rilascia anche chiudendo la connessione)
func (p *postgresDb) ReleaseLeadership(ctx context.Context) error {
	_, err := p.Lock.Exec(ctx, "SELECT pg_advisory_unlock($1)", environment.AggregationLockId)
	return err
}

// Close chiude la connessione al database PostgreSQL
func (p *postgresDb) Close() {
	p.Db.Close()
	err := p.Lock.Close(p.Ctx)
	if err != nil {
		logger.Log.Error("Unable to close the lock connection: ", err)
		os.Exit(1)
	}
}

// regionDB è l'istanza del database per i metadati della regione
var regionDB postgresDb = postgresDb{
	Db:   nil,
	Lock: nil,
	Ctx:  context.Background(),
	Url:  "",
}

// SetupRegionDbConnection configura e stabilisce la connessione al database dei metadati della regione
func SetupRegionDbConnection() error {

	// Se la connessione è già stabilita, non fare nulla
	if regionDB.Db != nil {
		logger.Log.Debug("Database connection already established")
		return nil
	}

	// Costruisce la stringa di connessione
	regionDB.Url = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		environment.PostgresRegionUser, environment.PostgresRegionPass, environment.PostgresRegionHost, environment.PostgresRegionPort, environment.PostgresRegionDatabase,
	)

	// Stabilisce la connessione
	return regionDB.Connect()
}

// TryAcquireAggregationLock prova ad acquisire il lock per l'aggregazione
func TryAcquireAggregationLock(ctx context.Context) (bool, error) {
	return regionDB.TryBecomeLeader(ctx)
}

// ReleaseAggregationLock rilascia il lock per l'aggregazione
func ReleaseAggregationLock(ctx context.Context) error {
	return regionDB.ReleaseLeadership(ctx)
}

// CloseRegionDbConnection chiude la connessione al database dei metadati della regione
func CloseRegionDbConnection() {
	if regionDB.Db != nil {
		regionDB.Close()
		logger.Log.Info("Region database connection closed")
	} else {
		logger.Log.Warn("Region database connection was not established")
	}
}

// sensorDB è l'istanza del database per le misurazioni dei sensori
var sensorDB postgresDb = postgresDb{
	Db:   nil,
	Lock: nil,
	Ctx:  context.Background(),
	Url:  "",
}

// SetupSensorDbConnection configura e stabilisce la connessione al database delle misurazioni dei sensori
func SetupSensorDbConnection() error {

	// Se la connessione è già stabilita, non fare nulla
	if sensorDB.Db != nil {
		logger.Log.Debug("Database connection already established")
		return nil
	}

	// Costruisce la stringa di connessione
	sensorDB.Url = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		environment.PostgresSensorUser, environment.PostgresSensorPass, environment.PostgresSensorHost, environment.PostgresSensorPort, environment.PostgresSensorDatabase,
	)

	// Stabilisce la connessione
	return sensorDB.Connect()
}

// CloseSensorDbConnection chiude la connessione al database delle misurazioni dei sensori
func CloseSensorDbConnection() {
	if sensorDB.Db != nil {
		sensorDB.Close()
		logger.Log.Info("Sensor database connection closed")
	} else {
		logger.Log.Warn("Sensor database connection was not established")
	}
}

// InsertSensorDataBatch inserisce un batch di dati dei sensori nel database gestendo i duplicati
func InsertSensorDataBatch(batch *types.SensorDataBatch) error {
	logger.Log.Info("Inserting sensor data batch")

	// Se il batch è vuoto, non fare nulla
	if batch.Count() == 0 {
		logger.Log.Info("No sensor data to insert, skipping")
		return nil
	}

	// Inizio transazione
	ctx := sensorDB.Ctx
	conn, err := sensorDB.Db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				logger.Log.Error("Unable to rollback transaction: ", err)
			} else {
				logger.Log.Debug("Transaction rolled back successfully")
			}
		} else {
			err := tx.Commit(ctx)
			if err != nil {
				logger.Log.Error("Unable to commit transaction: ", err)
			} else {
				logger.Log.Debug("Transaction committed successfully")
			}
		}
	}()

	// 1. Crea tabella temporanea
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_sensor_measurements (
			time TIMESTAMP,
			macrozone_name TEXT,
			zone_name TEXT,
			sensor_id TEXT,
			type TEXT,
			value DOUBLE PRECISION
		) ON COMMIT DROP;
	`)
	if err != nil {
		return err
	}

	// 2. Prepara i dati per l'inserimento
	rows := make([][]interface{}, 0, batch.Count())
	for _, d := range batch.Items() {
		timestamp := time.Unix(d.Timestamp, 0).UTC()
		rows = append(rows, []interface{}{
			timestamp,
			d.EdgeMacrozone,
			d.EdgeZone,
			d.SensorID,
			d.Type,
			d.Data,
		})
	}

	// 3. Inserisci i dati con CopyFrom nella tabella temporanea
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_sensor_measurements"},
		[]string{"time", "macrozone_name", "zone_name", "sensor_id", "type", "value"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return err
	}

	// 4. Copia nella tabella definitiva ignorando i duplicati
	_, err = tx.Exec(ctx, `
		INSERT INTO sensor_measurements (time, macrozone_name, zone_name, sensor_id, type, value)
		SELECT time, macrozone_name, zone_name, sensor_id, type, value FROM temp_sensor_measurements
		ON CONFLICT (time, macrozone_name, zone_name, sensor_id, type) DO NOTHING;
	`)
	if err != nil {
		return err
	}

	logger.Log.Info("Inserted sensor data batch successfully: ", len(batch.Items()), " entries")
	return nil
}

// UpdateSensorLastSeenBatch aggiorna il campo last_seen dei sensori in base ai dati ricevuti nel batch
func UpdateSensorLastSeenBatch(batch *types.SensorDataBatch) error {

	logger.Log.Info("Updating last seen batch")

	// Se il batch è vuoto, non fare nulla
	if batch.Count() == 0 {
		logger.Log.Info("No sensor data to update last seen, skipping")
		return nil
	}

	const keySeparator = "||"

	// Mappa per calcolare l'ultimo last_seen per ogni sensore
	// chiave = macrozone_name + keySeparator + zone_name + keySeparator + sensor_id
	lastSeenSensors := make(map[string]time.Time)

	// Calcola il timestamp massimo per ogni sensore
	for _, d := range batch.Items() {
		timestamp := time.Unix(d.Timestamp, 0).UTC()
		sensorKey := d.EdgeMacrozone + keySeparator + d.EdgeZone + keySeparator + d.SensorID
		if ts, ok := lastSeenSensors[sensorKey]; !ok || timestamp.After(ts) {
			lastSeenSensors[sensorKey] = timestamp
		}
	}

	// Funzione helper per creare slice [][]interface{} da map con chiave multipla splittata
	buildRows := func(m map[string]time.Time, parts int) [][]interface{} {
		rows := make([][]interface{}, 0, len(m))
		for k, ts := range m {
			ks := strings.Split(k, keySeparator)
			row := make([]interface{}, 0, parts+1)
			row = append(row, ts)
			for i := 0; i < parts; i++ {
				row = append(row, ks[i])
			}
			rows = append(rows, row)
		}
		return rows
	}

	// Inizio transazione
	ctx := regionDB.Ctx
	conn, err := regionDB.Db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	// Creo tabella temporanea per i last_seen dei sensori
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE tmp_sensors_last_seen (
			last_seen timestamptz,
			macrozone_name text,
			zone_name text,
			id text
		) ON COMMIT DROP
	`)
	if err != nil {
		return err
	}

	// Inserisco i dati calcolati nella tabella temporanea
	rows := buildRows(lastSeenSensors, 3)
	logger.Log.Debug("Updating last seen for sensors: ", len(rows), " entries")
	for r := range rows {
		logger.Log.Debug(" - ", rows[r])
	}
	_, err = tx.CopyFrom(ctx, pgx.Identifier{"tmp_sensors_last_seen"}, []string{"last_seen", "macrozone_name", "zone_name", "id"}, pgx.CopyFromRows(rows))
	if err != nil {
		return err
	}

	// Aggiorna la colonna last_seen nella tabella sensors
	_, err = tx.Exec(ctx, `
		UPDATE sensors s
		SET last_seen = tmp.max_last_seen
		FROM (
			SELECT macrozone_name, zone_name, id, MAX(last_seen) AS max_last_seen
			FROM tmp_sensors_last_seen
			GROUP BY macrozone_name, zone_name, id
		) tmp
		WHERE s.macrozone_name = tmp.macrozone_name
		  AND s.zone_name = tmp.zone_name
		  AND s.id = tmp.id
		  AND (s.last_seen IS NULL OR s.last_seen < tmp.max_last_seen)
		`)
	if err != nil {
		return err
	}

	// Aggiorno il last_seen del hub regionale
	err = UpdateLastSeenRegionHub()
	if err != nil {
		return err
	}

	return nil
}

// UpdateLastSeenRegionHub aggiorna il campo last_seen dell'hub regionale
func UpdateLastSeenRegionHub() error {
	query := `
		UPDATE region_hubs
		SET last_seen = NOW()
		WHERE id = $1
	`
	_, err := regionDB.Db.Exec(regionDB.Ctx, query, environment.HubID)
	return err
}

// InsertRegionStatisticsData inserisce i dati aggregati delle statistiche nel database
func InsertRegionStatisticsData(s types.AggregatedStats) error {
	query := `
		INSERT INTO region_aggregated_statistics (time, type, min_value, max_value, avg_value, avg_sum, avg_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	t := time.Unix(s.Timestamp, 0).UTC()
	_, err := sensorDB.Db.Exec(sensorDB.Ctx, query, t, s.Type, s.Min, s.Max, s.Avg, s.Sum, s.Count)
	return err
}

// GetMacrozoneStatisticsData esegue la query per ottenere le statistiche aggregate delle macrozone
func GetMacrozoneStatisticsData(ctx context.Context, startTime, endTime time.Time) ([]types.AggregatedStats, error) {
	query := `
		SELECT time, macrozone_name, type, min_value, max_value, avg_value, avg_sum, avg_count
		FROM macrozone_aggregated_statistics
		WHERE time >= $1 AND time < $2
	`
	rows, err := sensorDB.Db.Query(ctx, query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []types.AggregatedStats
	for rows.Next() {
		var s types.AggregatedStats
		var t time.Time
		err := rows.Scan(&t, &s.Macrozone, &s.Type, &s.Min, &s.Max, &s.Avg, &s.Sum, &s.Count)
		if err != nil {
			return nil, err
		}
		s.Timestamp = t.Unix()
		stats = append(stats, s)
	}
	return stats, nil
}

// GetLastRegionAggregatedData ritorna le ultime statistiche aggregate a livello di regione
func GetLastRegionAggregatedData(ctx context.Context) (types.AggregatedStats, error) {
	query := `
		SELECT 
		    time,
			type,
			min_value,
			max_value,
			avg_value,
			avg_sum,
			avg_count
		FROM region_aggregated_statistics
		ORDER BY time DESC
		LIMIT 1
	`
	rows, err := sensorDB.Db.Query(ctx, query)
	if err != nil {
		return types.AggregatedStats{}, fmt.Errorf("query for last macrozone aggregated data failed: %w", err)
	}
	defer rows.Close()

	var aggregatedStats types.AggregatedStats
	for rows.Next() {
		var t time.Time
		err := rows.Scan(&t, &aggregatedStats.Type, &aggregatedStats.Min, &aggregatedStats.Max, &aggregatedStats.Avg, &aggregatedStats.Sum, &aggregatedStats.Count)
		if err != nil {
			return types.AggregatedStats{}, fmt.Errorf("scanning last macrozone aggregated data failed: %w", err)
		}
		aggregatedStats.Timestamp = t.Unix()
	}

	return aggregatedStats, nil
}

// InsertMacrozoneStatisticsDataBatch inserisce i dati aggregati delle statistiche nel database in batch gestendo i duplicati
func InsertMacrozoneStatisticsDataBatch(batch *types.AggregatedStatsBatch) error {
	logger.Log.Info("Inserting macrozone statistics data batch")

	// Se il batch è vuoto, non fare nulla
	if batch.Count() == 0 {
		logger.Log.Info("No macrozone statistics data to insert, skipping")
		return nil
	}

	// Inizio transazione
	ctx := sensorDB.Ctx
	conn, err := sensorDB.Db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	// 1. Crea tabella temporanea
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_macrozone_aggregated_statistics (
			time TIMESTAMP,
			macrozone_name TEXT,
			type TEXT,
			min_value DOUBLE PRECISION,
			max_value DOUBLE PRECISION,
			avg_value DOUBLE PRECISION,
			avg_sum DOUBLE PRECISION,
			avg_count INT
		) ON COMMIT DROP;
	`)
	if err != nil {
		return err
	}

	// 2. Prepara i dati per l'inserimento
	rows := make([][]interface{}, 0, batch.Count())
	for _, s := range batch.Items() {
		timestamp := time.Unix(s.Timestamp, 0).UTC()
		rows = append(rows, []interface{}{
			timestamp,
			s.Macrozone,
			s.Type,
			s.Min,
			s.Max,
			s.Avg,
			s.Sum,
			s.Count,
		})
	}

	// 3. Inserisci i dati con CopyFrom nella tabella temporanea
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_macrozone_aggregated_statistics"},
		[]string{"time", "macrozone_name", "type", "min_value", "max_value", "avg_value", "avg_sum", "avg_count"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return err
	}

	// 4. Copia nella tabella definitiva ignorando i duplicati
	_, err = tx.Exec(ctx, `
		INSERT INTO macrozone_aggregated_statistics (time, macrozone_name, type, min_value, max_value, avg_value, avg_sum, avg_count)
		SELECT time, macrozone_name, type, min_value, max_value, avg_value, avg_sum, avg_count FROM temp_macrozone_aggregated_statistics
		ON CONFLICT (time, macrozone_name, type) DO NOTHING;
	`)
	if err != nil {
		return err
	}

	logger.Log.Info("Inserted macrozone statistics data batch successfully: ", len(batch.Items()), " entries")
	return nil
}

// InsertZoneStatisticsDataBatch inserisce i dati aggregati delle statistiche nel database in batch gestendo i duplicati
func InsertZoneStatisticsDataBatch(batch *types.AggregatedStatsBatch) error {
	logger.Log.Info("Inserting zone statistics data batch")

	// Se il batch è vuoto, non fare nulla
	if batch.Count() == 0 {
		logger.Log.Info("No zone statistics data to insert, skipping")
		return nil
	}

	// Inizio transazione
	ctx := sensorDB.Ctx
	conn, err := sensorDB.Db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	// 1. Crea tabella temporanea
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_zone_aggregated_statistics (
			time TIMESTAMP,
			macrozone_name TEXT,
			zone_name TEXT,
			type TEXT,
			min_value DOUBLE PRECISION,
			max_value DOUBLE PRECISION,
			avg_value DOUBLE PRECISION,
			avg_sum DOUBLE PRECISION,
			avg_count INT
		) ON COMMIT DROP;
	`)
	if err != nil {
		return err
	}

	// 2. Prepara i dati per l'inserimento
	rows := make([][]interface{}, 0, batch.Count())
	for _, s := range batch.Items() {
		timestamp := time.Unix(s.Timestamp, 0).UTC()
		rows = append(rows, []interface{}{
			timestamp,
			s.Macrozone,
			s.Zone,
			s.Type,
			s.Min,
			s.Max,
			s.Avg,
			s.Sum,
			s.Count,
		})
	}

	// 3. Inserisci i dati con CopyFrom nella tabella temporanea
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_zone_aggregated_statistics"},
		[]string{"time", "macrozone_name", "zone_name", "type", "min_value", "max_value", "avg_value", "avg_sum", "avg_count"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return err
	}

	// 4. Copia nella tabella definitiva ignorando i duplicati
	_, err = tx.Exec(ctx, `
		INSERT INTO zone_aggregated_statistics (time, macrozone_name, zone_name, type, min_value, max_value, avg_value, avg_sum, avg_count)
		SELECT time, macrozone_name, zone_name, type, min_value, max_value, avg_value, avg_sum, avg_count FROM temp_zone_aggregated_statistics
		ON CONFLICT (time, macrozone_name, zone_name, type) DO NOTHING;
	`)
	if err != nil {
		return err
	}

	logger.Log.Info("Inserted zone statistics data batch successfully: ", len(batch.Items()), " entries")
	return nil
}

// RegisterDevicesFromBatch registra o aggiorna hub e sensori in batch
func RegisterDevicesFromBatch(batch *types.ConfigurationMsgBatch) error {
	logger.Log.Info("Registering devices from batch")

	// Se il batch è vuoto, non fare nulla
	if batch.Count() == 0 {
		logger.Log.Info("No devices to register, skipping")
		return nil
	}

	// Inizio transazione
	ctx := regionDB.Ctx
	conn, err := regionDB.Db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				logger.Log.Error("Unable to rollback transaction: ", err)
				os.Exit(1)
			} else {
				logger.Log.Debug("Transaction rolled back successfully")
			}
		} else {
			err = tx.Commit(ctx)
			if err != nil {
				logger.Log.Error("Unable to commit transaction: ", err)
				os.Exit(1)
			} else {
				logger.Log.Debug("Transaction committed successfully")
			}
		}
	}()

	// 1. Crea tabelle temporanee
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE tmp_macrozone_hubs (
			hub_id TEXT,
			macrozone_name TEXT,
			service TEXT,
			timestamp TIMESTAMPTZ
		) ON COMMIT DROP;

		CREATE TEMP TABLE tmp_zone_hubs (
			hub_id TEXT,
			macrozone_name TEXT,
			zone_name TEXT,
			service TEXT,
			timestamp TIMESTAMPTZ
		) ON COMMIT DROP;

		CREATE TEMP TABLE tmp_sensors (
			sensor_id TEXT,
			macrozone_name TEXT,
			zone_name TEXT,
			sensor_type TEXT,
			sensor_reference TEXT,
			timestamp TIMESTAMPTZ
		) ON COMMIT DROP;
	`)
	if err != nil {
		return fmt.Errorf("errore creazione tabelle temporanee: %w", err)
	}

	// 2. Prepara i dati per l'inserimento
	now := time.Now().UTC()
	rowsMacro := make([][]interface{}, 0)
	rowsZone := make([][]interface{}, 0)
	rowsSensor := make([][]interface{}, 0)

	for _, msg := range batch.Items() {
		timestamp := time.Unix(msg.Timestamp, 0).UTC()
		if timestamp.IsZero() {
			timestamp = now
		}

		switch msg.MsgType {
		case types.NewSensorMsgType:
			// Nuovo sensore
			rowsSensor = append(rowsSensor, []interface{}{
				msg.SensorID, msg.EdgeMacrozone, msg.EdgeZone,
				msg.SensorType, msg.SensorReference, timestamp,
			})
		case types.NewEdgeMsgType:
			// Nuovo hub di zona
			rowsZone = append(rowsZone, []interface{}{
				msg.HubID, msg.EdgeMacrozone, msg.EdgeZone, msg.Service, timestamp,
			})
		case types.NewProximityMsgType:
			// Nuovo hub di macrozona
			rowsMacro = append(rowsMacro, []interface{}{
				msg.HubID, msg.EdgeMacrozone, msg.Service, timestamp,
			})
		default:
			// Messaggio non riconosciuto, salta
			logger.Log.Warn("Unknown message type in configuration batch: ", msg.MsgType)
			continue
		}
	}

	// 3. Inserisci i dati nelle temp tables
	if len(rowsMacro) > 0 {
		_, err = tx.CopyFrom(
			ctx,
			pgx.Identifier{"tmp_macrozone_hubs"},
			[]string{"hub_id", "macrozone_name", "service", "timestamp"},
			pgx.CopyFromRows(rowsMacro),
		)
		if err != nil {
			return fmt.Errorf("copyFrom tmp_macrozone_hubs: %w", err)
		}
	}
	if len(rowsZone) > 0 {
		_, err = tx.CopyFrom(
			ctx,
			pgx.Identifier{"tmp_zone_hubs"},
			[]string{"hub_id", "macrozone_name", "zone_name", "service", "timestamp"},
			pgx.CopyFromRows(rowsZone),
		)
		if err != nil {
			return fmt.Errorf("copyFrom tmp_zone_hubs: %w", err)
		}
	}
	if len(rowsSensor) > 0 {
		_, err = tx.CopyFrom(
			ctx,
			pgx.Identifier{"tmp_sensors"},
			[]string{"sensor_id", "macrozone_name", "zone_name", "sensor_type", "sensor_reference", "timestamp"},
			pgx.CopyFromRows(rowsSensor),
		)
		if err != nil {
			return fmt.Errorf("copyFrom tmp_sensors: %w", err)
		}
	}

	// 4. Copia nelle tabelle reali con upsert
	_, err = tx.Exec(ctx, `
		INSERT INTO macrozone_hubs (id, macrozone_name, service, registration_time, last_seen)
		SELECT hub_id,
			   macrozone_name,
			   service,
			   MIN(timestamp) AS registration_time,
			   MAX(timestamp) AS last_seen
		FROM tmp_macrozone_hubs
		GROUP BY hub_id, macrozone_name, service
		ON CONFLICT (id, macrozone_name) DO
		UPDATE SET last_seen = EXCLUDED.last_seen
		WHERE macrozone_hubs.last_seen IS NULL OR macrozone_hubs.last_seen < EXCLUDED.last_seen;
		
		INSERT INTO zone_hubs (id, macrozone_name, zone_name, service, registration_time, last_seen)
		SELECT hub_id,
			   macrozone_name,
			   zone_name,
			   service,
			   MIN(timestamp) AS registration_time,
			   MAX(timestamp) AS last_seen
		FROM tmp_zone_hubs
		GROUP BY hub_id, macrozone_name, zone_name, service
		ON CONFLICT (id, macrozone_name, zone_name) DO
		UPDATE SET last_seen = EXCLUDED.last_seen
		WHERE zone_hubs.last_seen IS NULL OR zone_hubs.last_seen < EXCLUDED.last_seen;
		
		INSERT INTO sensors (id, macrozone_name, zone_name, type, reference, registration_time, last_seen)
		SELECT sensor_id,
			   macrozone_name,
			   zone_name,
			   sensor_type,
			   sensor_reference,
			   MIN(timestamp) AS registration_time,
			   MAX(timestamp) AS last_seen
		FROM tmp_sensors
		GROUP BY sensor_id, macrozone_name, zone_name, sensor_type, sensor_reference
		ON CONFLICT (id, macrozone_name, zone_name) DO
		UPDATE SET last_seen = EXCLUDED.last_seen
		WHERE sensors.last_seen IS NULL OR sensors.last_seen < EXCLUDED.last_seen;
	`)
	if err != nil {
		return fmt.Errorf("errore insert finali: %w", err)
	}

	logger.Log.Info("Registered devices batch successfully: ", len(batch.Items()), " entries")
	return nil
}

// SelfRegistration registra o aggiorna l'hub regionale
func SelfRegistration() error {

	// Assicura che la connessione al database sia attiva
	err := SetupRegionDbConnection()
	if err != nil {
		logger.Log.Error("Failed to connect to the sensor database: ", err)
		os.Exit(1)
	}

	// Inserisce o aggiorna l'hub regionale
	query := `
	INSERT INTO region_hubs (id, service, registration_time, last_seen)
	VALUES ($1, $2, CURRENT_TIMESTAMP AT TIME ZONE 'UTC', CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
	ON CONFLICT (id) DO UPDATE
	SET last_seen = EXCLUDED.last_seen
	WHERE region_hubs.last_seen IS NULL OR region_hubs.last_seen < EXCLUDED.last_seen
`
	_, err = regionDB.Db.Exec(regionDB.Ctx, query, environment.HubID, environment.ServiceMode)
	return err
}

// UpdateHubLastSeen aggiorna il campo last_seen di hub e sensori in base ai messaggi di heartbeat ricevuti
func UpdateHubLastSeen(batch *types.HeartbeatMsgBatch) error {
	logger.Log.Info("Updating hub last_seen from heartbeat batch")

	// Se il batch è vuoto, non fare nulla
	if batch.Count() == 0 {
		logger.Log.Info("No heartbeat data to update, skipping")
		return nil
	}

	// Inizio transazione
	ctx := regionDB.Ctx
	conn, err := regionDB.Db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				logger.Log.Error("Unable to rollback transaction: ", err)
				os.Exit(1)
			} else {
				logger.Log.Debug("Transaction rolled back successfully")
			}
		} else {
			err = tx.Commit(ctx)
			if err != nil {
				logger.Log.Error("Unable to commit transaction: ", err)
				os.Exit(1)
			} else {
				logger.Log.Debug("Transaction committed successfully")
			}
		}
	}()

	// 1. Crea tabelle temporanee
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE tmp_macrozone_hb (
			hub_id TEXT,
			macrozone_name TEXT,
			timestamp TIMESTAMPTZ
		) ON COMMIT DROP;

		CREATE TEMP TABLE tmp_zone_hb (
			hub_id TEXT,
			macrozone_name TEXT,
			zone_name TEXT,
			timestamp TIMESTAMPTZ
		) ON COMMIT DROP;
	`)
	if err != nil {
		return fmt.Errorf("errore creazione tabelle temporanee: %w", err)
	}

	// 2. Prepara i dati per l’inserimento
	now := time.Now().UTC()
	rowsMacro := make([][]interface{}, 0)
	rowsZone := make([][]interface{}, 0)

	for _, hb := range batch.Items() {
		timestamp := time.Unix(hb.Timestamp, 0).UTC()
		if timestamp.IsZero() {
			timestamp = now
		}

		if hb.HubID == "" {
			// Messaggio non valido, salta
			logger.Log.Warn("Heartbeat message with empty HubID, skipping")
			continue
		}

		if hb.EdgeMacrozone == "" {
			// Messaggio non valido, salta
			logger.Log.Warn("Heartbeat message with empty Macrozone, skipping")
			continue
		}

		if hb.EdgeZone != "" {
			rowsZone = append(rowsZone, []interface{}{hb.HubID, hb.EdgeMacrozone, hb.EdgeZone, timestamp})
		} else {
			rowsMacro = append(rowsMacro, []interface{}{hb.HubID, hb.EdgeMacrozone, timestamp})
		}
	}

	// 3. Inserisci i dati nelle temp tables
	if len(rowsMacro) > 0 {
		_, err = tx.CopyFrom(
			ctx,
			pgx.Identifier{"tmp_macrozone_hb"},
			[]string{"hub_id", "macrozone_name", "timestamp"},
			pgx.CopyFromRows(rowsMacro),
		)
		if err != nil {
			return fmt.Errorf("copyFrom tmp_macrozone_hb: %w", err)
		}
	}
	if len(rowsZone) > 0 {
		_, err = tx.CopyFrom(
			ctx,
			pgx.Identifier{"tmp_zone_hb"},
			[]string{"hub_id", "macrozone_name", "zone_name", "timestamp"},
			pgx.CopyFromRows(rowsZone),
		)
		if err != nil {
			return fmt.Errorf("copyFrom tmp_zone_hb: %w", err)
		}
	}

	// 4. Update massivo nelle tabelle reali
	_, err = tx.Exec(ctx, `
		UPDATE macrozone_hubs m
		SET last_seen = t.timestamp
		FROM tmp_macrozone_hb t
		WHERE m.id = t.hub_id
		  AND m.macrozone_name = t.macrozone_name
		  AND (m.last_seen IS NULL OR m.last_seen < t.timestamp);

		UPDATE zone_hubs z
		SET last_seen = t.timestamp
		FROM tmp_zone_hb t
		WHERE z.id = t.hub_id
		  AND z.macrozone_name = t.macrozone_name
		  AND z.zone_name = t.zone_name
		  AND (z.last_seen IS NULL OR z.last_seen < t.timestamp);
	`)
	if err != nil {
		return fmt.Errorf("errore update last_seen: %w", err)
	}

	logger.Log.Info("Updated hub last_seen successfully: ", len(batch.Items()), " entries")
	return nil
}
