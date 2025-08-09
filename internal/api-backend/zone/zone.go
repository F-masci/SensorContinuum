package zone

import (
	"SensorContinuum/internal/api-backend/storage"
	"SensorContinuum/pkg/types"
	"context"
	"database/sql"
	"errors"
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
