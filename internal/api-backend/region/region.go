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
