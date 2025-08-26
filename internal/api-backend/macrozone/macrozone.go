package macrozone

import (
	"SensorContinuum/internal/api-backend/storage"
	"SensorContinuum/internal/api-backend/zone"
	"SensorContinuum/pkg/types"
	"context"
	"database/sql"
	"errors"
	"time"
)

// GetMacrozonesList Restituisce la lista delle macrozone per una regione, con il conteggio delle zone associate
func GetMacrozonesList(ctx context.Context, regionName string) ([]types.Macrozone, error) {
	db, err := storage.GetCloudPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := db.Conn().Query(ctx, `
		SELECT m.region_name, m.name,
			ST_Y(m.location) as lat, ST_X(m.location) as lon,
			m.creation_time,
			COUNT(z.name) as zone_count
		FROM macrozones m
		LEFT JOIN zones z ON z.region_name = m.region_name AND z.macrozone_name = m.name
		WHERE m.region_name = $1
		GROUP BY m.region_name, m.name, m.location, m.creation_time
	`, regionName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var macrozones []types.Macrozone
	for rows.Next() {
		var m types.Macrozone
		if err := rows.Scan(&m.RegionName, &m.Name, &m.Lat, &m.Lon, &m.CreationTime, &m.ZoneCount); err != nil {
			return nil, err
		}
		macrozones = append(macrozones, m)
	}
	return macrozones, nil
}

// GetMacrozoneByName Restituisce una macrozona per nome, con la lista delle zone, hub, hub di zona e sensori associati
func GetMacrozoneByName(ctx context.Context, regionName string, name string) (*types.Macrozone, error) {
	cloudDb, err := storage.GetCloudPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	var m types.Macrozone
	err = cloudDb.Conn().QueryRow(ctx, `
		SELECT m.region_name, m.name,
			ST_Y(m.location) as lat, ST_X(m.location) as lon,
			m.creation_time
		FROM macrozones m
		WHERE m.region_name = $1 AND m.name = $2
	`, regionName, name).Scan(&m.RegionName, &m.Name, &m.Lat, &m.Lon, &m.CreationTime)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Carica le zone associate
	zones, err := zone.GetZonesList(ctx, m.RegionName, m.Name)
	if err != nil {
		return nil, err
	}
	m.Zones = zones
	m.ZoneCount = len(zones)

	regionDb, err := storage.GetRegionPostgresDB(ctx, regionName)
	if err != nil {
		return nil, err
	}

	// Carica gli hub di macrozona
	hubRows, err := regionDb.Conn().Query(ctx, `
		SELECT id, macrozone_name, service, registration_time, last_seen
		FROM macrozone_hubs
		WHERE macrozone_name = $1
	`, name)
	if err != nil {
		return nil, err
	}
	defer hubRows.Close()
	m.Hubs = make([]types.MacrozoneHub, 0)
	for hubRows.Next() {
		var hub types.MacrozoneHub
		if err := hubRows.Scan(&hub.Id, &hub.MacrozoneName, &hub.Service, &hub.RegistrationTime, &hub.LastSeen); err != nil {
			return nil, err
		}
		m.Hubs = append(m.Hubs, hub)
	}

	// Carica gli hub di zona
	zoneHubRows, err := regionDb.Conn().Query(ctx, `
		SELECT id, macrozone_name, zone_name, service, registration_time, last_seen
		FROM zone_hubs
		WHERE macrozone_name = $1
	`, name)
	if err != nil {
		return nil, err
	}
	defer zoneHubRows.Close()
	m.ZoneHubs = make([]types.ZoneHub, 0)
	for zoneHubRows.Next() {
		var zh types.ZoneHub
		if err := zoneHubRows.Scan(&zh.Id, &zh.MacrozoneName, &zh.ZoneName, &zh.Service, &zh.RegistrationTime, &zh.LastSeen); err != nil {
			return nil, err
		}
		m.ZoneHubs = append(m.ZoneHubs, zh)
	}

	// Carica i sensori associati alla macrozona
	sensorRows, err := regionDb.Conn().Query(ctx, `
		SELECT id, macrozone_name, zone_name, type, reference, registration_time, last_seen
		FROM sensors
		WHERE macrozone_name = $1
	`, name)
	if err != nil {
		return nil, err
	}
	defer sensorRows.Close()
	m.Sensors = make([]types.Sensor, 0)
	for sensorRows.Next() {
		var s types.Sensor
		if err := sensorRows.Scan(&s.Id, &s.MacrozoneName, &s.ZoneName, &s.Type, &s.Reference, &s.RegistrationTime, &s.LastSeen); err != nil {
			return nil, err
		}
		m.Sensors = append(m.Sensors, s)
	}

	return &m, nil
}

// GetAggregatedSensorData Restituisce i dati aggregati di una macrozona
func GetAggregatedSensorData(ctx context.Context, regionName, macrozoneName string, limit int) (*[]types.AggregatedStats, error) {
	sensorDb, err := storage.GetSensorPostgresDB(ctx, regionName)
	if err != nil {
		return nil, err
	}
	var a []types.AggregatedStats
	aggregatedDataRows, err := sensorDb.Conn().Query(ctx, `
		SELECT m.time, m.macrozone_name, m.type, m.min_value, m.max_value, m.avg_value
		FROM macrozone_aggregated_statistics m
		WHERE m.macrozone_name = $1
		ORDER BY m.time DESC
		LIMIT $2
	`, macrozoneName, limit)
	if err != nil {
		return nil, err
	}
	defer aggregatedDataRows.Close()
	a = make([]types.AggregatedStats, 0)
	for aggregatedDataRows.Next() {
		var ts time.Time
		var as types.AggregatedStats
		if err := aggregatedDataRows.Scan(&ts, &as.Macrozone, &as.Type, &as.Min, &as.Max, &as.Avg); err != nil {
			return nil, err
		}
		as.Timestamp = ts.Unix()
		a = append(a, as)
	}

	return &a, nil
}
