package zone

import (
	"SensorContinuum/internal/api-backend/storage"
	"SensorContinuum/pkg/types"
	"context"
	"database/sql"
	"errors"
	"time"
)

// GetZonesList Restituisce la lista delle zone per una macrozona
func GetZonesList(ctx context.Context, regionName, macrozoneName string) ([]types.Zone, error) {
	db, err := storage.GetCloudPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := db.Conn().Query(ctx, `
		SELECT z.region_name, z.macrozone_name, z.name, z.creation_time
		FROM zones z
		WHERE z.region_name = $1 AND z.macrozone_name = $2
	`, regionName, macrozoneName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []types.Zone
	for rows.Next() {
		var z types.Zone
		if err := rows.Scan(&z.RegionName, &z.MacrozoneName, &z.Name, &z.CreationTime); err != nil {
			return nil, err
		}
		zones = append(zones, z)
	}
	return zones, nil
}

// GetZoneByName Restituisce una zona per nome, con la lista degli hub e dei sensori associati
func GetZoneByName(ctx context.Context, regionName, macrozoneName, name string) (*types.Zone, error) {
	cloudDb, err := storage.GetCloudPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	var z types.Zone
	err = cloudDb.Conn().QueryRow(ctx, `
		SELECT z.region_name, z.macrozone_name, z.name, z.creation_time
		FROM zones z
		WHERE z.region_name = $1 AND z.macrozone_name = $2 AND z.name = $3
	`, regionName, macrozoneName, name).Scan(&z.RegionName, &z.MacrozoneName, &z.Name, &z.CreationTime)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	regionDb, err := storage.GetRegionPostgresDB(ctx, regionName)
	if err != nil {
		return nil, err
	}

	// Carica gli hub di zona
	zoneHubRows, err := regionDb.Conn().Query(ctx, `
		SELECT id, macrozone_name, zone_name, service, registration_time, last_seen
		FROM zone_hubs
		WHERE macrozone_name = $1 AND zone_name = $2
	`, macrozoneName, name)
	if err != nil {
		return nil, err
	}
	defer zoneHubRows.Close()
	z.Hubs = make([]types.ZoneHub, 0)
	for zoneHubRows.Next() {
		var zh types.ZoneHub
		if err := zoneHubRows.Scan(&zh.Id, &zh.MacrozoneName, &zh.ZoneName, &zh.Service, &zh.RegistrationTime, &zh.LastSeen); err != nil {
			return nil, err
		}
		z.Hubs = append(z.Hubs, zh)
	}

	// Carica i sensori associati alla zona
	sensorRows, err := regionDb.Conn().Query(ctx, `
		SELECT id, macrozone_name, zone_name, type, reference, registration_time, last_seen
		FROM sensors
		WHERE macrozone_name = $1 AND zone_name = $2
	`, macrozoneName, name)
	if err != nil {
		return nil, err
	}
	defer sensorRows.Close()
	z.Sensors = make([]types.Sensor, 0)
	for sensorRows.Next() {
		var s types.Sensor
		if err := sensorRows.Scan(&s.Id, &s.MacrozoneName, &s.ZoneName, &s.Type, &s.Reference, &s.RegistrationTime, &s.LastSeen); err != nil {
			return nil, err
		}
		z.Sensors = append(z.Sensors, s)
	}

	return &z, nil
}

// GetRawSensorData Restituisce i dati grezzi di un sensore in una zona
func GetRawSensorData(ctx context.Context, regionName, macrozoneName, zoneName, sensorId string, limit int) (*[]types.SensorData, error) {
	sensorDb, err := storage.GetSensorPostgresDB(ctx, regionName)
	if err != nil {
		return nil, err
	}
	var s []types.SensorData
	sensorDataRows, err := sensorDb.Conn().Query(ctx, `
		SELECT s.time, s.macrozone_name, s.zone_name, s.sensor_id, s.type, s.value
		FROM sensor_measurements s
		WHERE s.macrozone_name = $1 AND s.zone_name = $2 AND s.sensor_id = $3
		ORDER BY s.time DESC
		LIMIT $4
	`, macrozoneName, zoneName, sensorId, limit)
	if err != nil {
		return nil, err
	}
	defer sensorDataRows.Close()
	s = make([]types.SensorData, 0)
	for sensorDataRows.Next() {
		var ts time.Time
		var sd types.SensorData
		if err := sensorDataRows.Scan(&ts, &sd.EdgeMacrozone, &sd.EdgeZone, &sd.SensorID, &sd.Type, &sd.Data); err != nil {
			return nil, err
		}
		sd.Timestamp = ts.Unix()
		s = append(s, sd)
	}

	return &s, nil
}

// GetAggregatedSensorData Restituisce i dati aggregati di una zona
func GetAggregatedSensorData(ctx context.Context, regionName, macrozoneName, zoneName string, limit int) (*[]types.AggregatedStats, error) {
	sensorDb, err := storage.GetSensorPostgresDB(ctx, regionName)
	if err != nil {
		return nil, err
	}
	var a []types.AggregatedStats
	aggregatedDataRows, err := sensorDb.Conn().Query(ctx, `
		SELECT z.time, z.macrozone_name, z.zone_name, z.type, z.min_value, z.max_value, z.avg_value
		FROM zone_aggregated_statistics z
		WHERE z.macrozone_name = $1 AND z.zone_name = $2
		ORDER BY z.time DESC
		LIMIT $3
	`, macrozoneName, zoneName, limit)
	if err != nil {
		return nil, err
	}
	defer aggregatedDataRows.Close()
	a = make([]types.AggregatedStats, 0)
	for aggregatedDataRows.Next() {
		var ts time.Time
		var as types.AggregatedStats
		if err := aggregatedDataRows.Scan(&ts, &as.Macrozone, &as.Zone, &as.Type, &as.Min, &as.Max, &as.Avg); err != nil {
			return nil, err
		}
		as.Timestamp = ts.Unix()
		a = append(a, as)
	}

	return &a, nil
}
