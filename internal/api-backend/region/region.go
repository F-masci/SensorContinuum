package region

import (
	"SensorContinuum/internal/api-backend/macrozone"
	"SensorContinuum/internal/api-backend/storage"
	"SensorContinuum/pkg/types"
	"context"
	"database/sql"
	"errors"
	"time"
)

func GetAllRegions(ctx context.Context) ([]types.Region, error) {
	db, err := storage.GetCloudPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := db.Conn().Query(ctx, `
		SELECT r.name, COUNT(m.name) AS macrozone_count
		FROM regions r
		LEFT JOIN macrozones m ON m.region_name = r.name
		GROUP BY r.name
		ORDER BY r.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []types.Region
	for rows.Next() {
		var r types.Region
		if err := rows.Scan(&r.Name, &r.MacrozoneCount); err != nil {
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}

func GetRegionByName(ctx context.Context, name string) (*types.Region, error) {
	cloudDb, err := storage.GetCloudPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	var r types.Region
	err = cloudDb.Conn().QueryRow(ctx, `SELECT r.name FROM regions r WHERE r.name = $1`, name).Scan(&r.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Carica le macrozone associate
	macrozones, err := macrozone.GetMacrozonesList(ctx, r.Name)
	if err != nil {
		return nil, err
	}
	r.Macrozones = macrozones
	r.MacrozoneCount = len(macrozones)

	// Carica i region_hub associati
	regionDb, err := storage.GetRegionPostgresDB(ctx, name)
	if err != nil {
		return nil, err
	}
	rows, err := regionDb.Conn().Query(ctx, `
		SELECT id, service, registration_time, last_seen
		FROM region_hubs
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	r.Hubs = make([]types.RegionHub, 0)
	for rows.Next() {
		var hub types.RegionHub
		if err := rows.Scan(&hub.Id, &hub.Service, &hub.RegistrationTime, &hub.LastSeen); err != nil {
			return nil, err
		}
		r.Hubs = append(r.Hubs, hub)
	}

	return &r, nil
}

// GetAggregatedSensorData Restituisce i dati aggregati di una regione
func GetAggregatedSensorData(ctx context.Context, regionName string, limit int) (*[]types.AggregatedStats, error) {
	sensorDb, err := storage.GetSensorPostgresDB(ctx, regionName)
	if err != nil {
		return nil, err
	}
	var a []types.AggregatedStats
	aggregatedDataRows, err := sensorDb.Conn().Query(ctx, `
		SELECT m.time, m.type, m.min_value, m.max_value, m.avg_value
		FROM region_aggregated_statistics m
		ORDER BY m.time DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer aggregatedDataRows.Close()
	a = make([]types.AggregatedStats, 0)
	for aggregatedDataRows.Next() {
		var ts time.Time
		var as types.AggregatedStats
		if err := aggregatedDataRows.Scan(&ts, &as.Type, &as.Min, &as.Max, &as.Avg); err != nil {
			return nil, err
		}
		as.Timestamp = ts.Unix()
		a = append(a, as)
	}

	return &a, nil
}

// ComputeAggregatedSensorData Calcola i dati aggregati di una regione in un intervallo di tempo
func ComputeAggregatedSensorData(ctx context.Context, regionName string, start time.Time, end time.Time) ([]types.AggregatedStats, error) {
	sensorDb, err := storage.GetSensorPostgresDB(ctx, regionName)
	if err != nil {
		return nil, err
	}
	var a []types.AggregatedStats
	aggregatedDataRows, err := sensorDb.Conn().Query(ctx, `
		SELECT 
			type,
			MIN(min_value) as min_val,
			MAX(max_value) as max_val,
			CASE 
				WHEN SUM(avg_count) = 0 THEN 0
				ELSE SUM(avg_sum) / SUM(avg_count)
			END as avg_val,
			SUM(avg_sum) as avg_sum_val,
			COUNT(avg_count) as avg_count_val
		FROM macrozone_aggregated_statistics
		WHERE time >= $1 AND time < $2
		GROUP BY type
	`, start, end)
	if err != nil {
		return nil, err
	}
	defer aggregatedDataRows.Close()
	a = make([]types.AggregatedStats, 0)
	for aggregatedDataRows.Next() {
		var as types.AggregatedStats
		if err := aggregatedDataRows.Scan(&as.Type, &as.Min, &as.Max, &as.Avg, &as.Sum, &as.Count); err != nil {
			return nil, err
		}
		as.Timestamp = end.Unix()
		as.Region = regionName
		a = append(a, as)
	}

	return a, nil
}

// SaveAggregatedSensorData Salva i dati aggregati di una regione
func SaveAggregatedSensorData(ctx context.Context, regionName string, data types.AggregatedStats, timestamp time.Time) error {
	sensorDb, err := storage.GetSensorPostgresDB(ctx, regionName)
	if err != nil {
		return err
	}

	_, err = sensorDb.Conn().Exec(ctx, `
		INSERT INTO region_aggregated_statistics (time, type, min_value, max_value, avg_value, avg_sum, avg_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, timestamp, data.Type, data.Min, data.Max, data.Avg, data.Sum, data.Count)
	return err
}
