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

type postgresDb struct {
	Db  *pgxpool.Pool
	Ctx context.Context
	Url string
}

func (p *postgresDb) Connect() error {

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
	logger.Log.Info("Connected to the database successfully")
	return nil
}

func (p *postgresDb) Close() {
	p.Db.Close()
}

var regionDB postgresDb = postgresDb{
	Db:  nil,
	Ctx: context.Background(),
	Url: "",
}

func SetupRegionDbConnection() error {

	if regionDB.Db != nil {
		logger.Log.Debug("Database connection already established")
		return nil
	}

	regionDB.Url = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		environment.PostgresRegionUser, environment.PostgresRegionPass, environment.PostgresRegionHost, environment.PostgresRegionPort, environment.PostgresRegionDatabase,
	)

	return regionDB.Connect()
}

func CloseRegionDbConnection() {
	if regionDB.Db != nil {
		regionDB.Close()
		logger.Log.Info("Region database connection closed")
	} else {
		logger.Log.Warn("Region database connection was not established")
	}
}

var sensorDB postgresDb = postgresDb{
	Db:  nil,
	Ctx: context.Background(),
	Url: "",
}

func SetupSensorDbConnection() error {

	if sensorDB.Db != nil {
		logger.Log.Debug("Database connection already established")
		return nil
	}

	sensorDB.Url = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		environment.PostgresSensorUser, environment.PostgresSensorPass, environment.PostgresSensorHost, environment.PostgresSensorPort, environment.PostgresSensorDatabase,
	)

	return sensorDB.Connect()
}

func CloseSensorDbConnection() {
	if sensorDB.Db != nil {
		sensorDB.Close()
		logger.Log.Info("Sensor database connection closed")
	} else {
		logger.Log.Warn("Sensor database connection was not established")
	}
}

func InsertSensorDataBatch(batch types.SensorDataBatch) error {

	if batch.Count() == 0 {
		logger.Log.Warn("No sensor data to insert")
		return nil
	}

	tableName := pgx.Identifier{"sensor_measurements"}
	columns := []string{"time", "macrozone_name", "zone_name", "sensor_id", "type", "value"}

	rows := make([][]interface{}, 0, batch.Count())
	for _, d := range batch.SensorData {
		timestamp, err := time.Parse(time.RFC3339, d.Timestamp)
		if err != nil {
			logger.Log.Error("Failed to parse timestamp: ", err)
			return err
		}
		rows = append(rows, []interface{}{
			timestamp,
			d.EdgeMacrozone,
			d.EdgeZone,
			d.SensorID,
			d.Type,
			d.Data,
		})
	}

	count, err := sensorDB.Db.CopyFrom(
		sensorDB.Ctx,
		tableName,
		columns,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return err
	}

	logger.Log.Info("Inserted sensor data rows successfully: ", count)
	return nil
}

func UpdateLastSeenBatch(batch types.SensorDataBatch) error {

	if batch.Count() == 0 {
		logger.Log.Warn("No sensor data to update last seen")
		return nil
	}

	const keySeparator = "||"

	// Mappa per calcolare l'ultimo last_seen per ogni sensore
	// chiave = macrozone_name + keySeparator + zone_name + keySeparator + sensor_id
	lastSeenSensors := make(map[string]time.Time)

	// Calcola il timestamp massimo per ogni sensore
	for _, d := range batch.SensorData {
		timestamp, err := time.Parse(time.RFC3339, d.Timestamp)
		if err != nil {
			logger.Log.Error("Failed to parse timestamp: ", err)
			return err
		}

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

	rows := buildRows(lastSeenSensors, 3)
	_, err = tx.CopyFrom(ctx, pgx.Identifier{"tmp_sensors_last_seen"}, []string{"last_seen", "macrozone_name", "zone_name", "id"}, pgx.CopyFromRows(rows))
	if err != nil {
		return err
	}

	// Aggiorna la colonna last_seen nella tabella sensors
	_, err = tx.Exec(ctx, `
		UPDATE sensors s
		SET last_seen = tmp.last_seen
		FROM tmp_sensors_last_seen tmp
		WHERE s.macrozone_name = tmp.macrozone_name
		  AND s.zone_name = tmp.zone_name
		  AND s.id = tmp.id
		  AND (s.last_seen IS NULL OR s.last_seen < tmp.last_seen)
	`)
	if err != nil {
		return err
	}

	// Aggiorno il last_seen del hub regionale
	_, err = tx.Exec(ctx, `
		UPDATE region_hubs rh
		SET last_seen = NOW()
		WHERE rh.id = $1
	`, environment.HubID)
	if err != nil {
		return err
	}

	return nil
}

// InsertStatisticsData inserisce i dati aggregati delle statistiche nel database
func InsertStatisticsData(s types.AggregatedStats) error {
	query := `
        INSERT INTO aggregated_statistics (time, macrozone_name, type, min_value, max_value, avg_value)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := sensorDB.Db.Exec(sensorDB.Ctx, query, s.Timestamp, s.Macrozone, s.Type, s.Min, s.Max, s.Avg)
	return err
}

// RegisterMacrozoneHub Registra o aggiorna un hub di macrozona (proximity fog hub)
func RegisterMacrozoneHub(msg types.ConfigurationMsg) error {
	timestamp := time.Unix(msg.Timestamp, 0)
	query := `
		INSERT INTO macrozone_hubs (id, macrozone_name, service, registration_time, last_seen)
		VALUES ($1, $2, $3, $4, $4)
		ON CONFLICT (id, macrozone_name) DO UPDATE SET last_seen = EXCLUDED.last_seen
	`
	_, err := regionDB.Db.Exec(regionDB.Ctx, query, msg.HubID, msg.EdgeMacrozone, msg.Service, timestamp)
	return err
}

// RegisterZoneHub Registra o aggiorna un hub di zona (edge hub)
func RegisterZoneHub(msg types.ConfigurationMsg) error {
	timestamp := time.Unix(msg.Timestamp, 0)
	query := `
		INSERT INTO zone_hubs (id, macrozone_name, zone_name, service, registration_time, last_seen)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (id, macrozone_name, zone_name) DO UPDATE SET last_seen = EXCLUDED.last_seen
	`
	_, err := regionDB.Db.Exec(regionDB.Ctx, query, msg.HubID, msg.EdgeMacrozone, msg.EdgeZone, msg.Service, timestamp)
	return err
}

// RegisterSensor Registra o aggiorna un sensore associato a un edge hub
func RegisterSensor(msg types.ConfigurationMsg) error {
	timestamp := time.Unix(msg.Timestamp, 0)
	query := `
        INSERT INTO sensors (id, macrozone_name, zone_name, type, reference, registration_time, last_seen)
        VALUES ($1, $2, $3, $4, $5, $6, $6)
        ON CONFLICT (id, macrozone_name, zone_name) DO UPDATE SET last_seen = EXCLUDED.last_seen
    `
	_, err := regionDB.Db.Exec(regionDB.Ctx, query, msg.SensorID, msg.EdgeMacrozone, msg.EdgeZone, msg.SensorType, msg.SensorReference, timestamp)
	return err
}

func Register() error {

	err := SetupRegionDbConnection()
	if err != nil {
		logger.Log.Error("Failed to connect to the sensor database: ", err)
		os.Exit(1)
	}

	query := `
	INSERT INTO region_hubs (id, service, registration_time, last_seen)
	VALUES ($1, $2, NOW(), NOW())
	ON CONFLICT (id) DO UPDATE SET last_seen = EXCLUDED.last_seen
`
	_, err = regionDB.Db.Exec(regionDB.Ctx, query, environment.HubID, types.IntrermediateHubService)
	return err
}
