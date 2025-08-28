package macrozone

import (
	"SensorContinuum/internal/api-backend/storage"
	"SensorContinuum/pkg/types"
	"context"
)

// GetMacrozoneByLocation restituisce le macrozone che contengono o sono entro un certo raggio dalla posizione specificata (lat, lon).
// La funzione utilizza PostGIS per eseguire query spaziali sul database PostgreSQL.
func GetMacrozoneByLocation(ctx context.Context, lat, lon, radius float64) ([]types.Macrozone, error) {

	// 1. Connettersi al cloudDB per leggere tutte le macrozone vicine
	cloudDb, err := storage.GetCloudPostgresDB(ctx)
	if err != nil {
		return nil, err
	}

	queryMacrozones := `
		SELECT region_name, name
		FROM macrozones
		WHERE ST_Contains(
				location,
				ST_SetSRID(ST_MakePoint($1, $2), 4326)
			)
		   OR ST_DWithin(
				location::geography,
				ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
				$3
			)
	`
	rows, err := cloudDb.Conn().Query(ctx, queryMacrozones, lon, lat, radius)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 2. Raggruppa le macrozone per regione
	regions := make([]types.Macrozone, 0)
	for rows.Next() {
		var m types.Macrozone
		if err := rows.Scan(&m.RegionName, &m.Name); err != nil {
			return nil, err
		}
		regions = append(regions, m)
	}

	return regions, nil
}

// GetMacrozoneNeighbors restituisce le macrozone confinanti con la macrozona specificata.
func GetMacrozoneNeighbors(ctx context.Context, macrozoneName string, radius float64) ([]types.Macrozone, error) {

	cloudDb, err := storage.GetCloudPostgresDB(ctx)
	if err != nil {
		return nil, err
	}

	queryMacrozones := `
		WITH macrozone_centroids AS (
			-- centri delle macrozone
			SELECT 
				region_name,
				name,
				ST_Centroid(location) AS center
			FROM macrozones
		)
		SELECT b.region_name, b.name AS neighbor
		FROM macrozone_centroids a
		JOIN macrozone_centroids b
		  ON a.name <> b.name
		 AND ST_DWithin(a.center::geography, b.center::geography, $2) -- distanza in metri
		WHERE a.name = $1
	`
	rows, err := cloudDb.Conn().Query(ctx, queryMacrozones, macrozoneName, radius)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	regions := make([]types.Macrozone, 0)
	for rows.Next() {
		var m types.Macrozone
		if err := rows.Scan(&m.RegionName, &m.Name); err != nil {
			return nil, err
		}
		regions = append(regions, m)
	}

	return regions, nil
}

// GetAllMacrozoneNeighbors richiama GetMacrozoneNeighbors per tutte le macrozone
// e costruisce una mappa macrozona -> vicini.
func GetAllMacrozoneNeighbors(ctx context.Context, macrozones []types.Macrozone, radius float64) (map[string][]types.Macrozone, error) {
	neighborsMap := make(map[string][]types.Macrozone)

	for _, m := range macrozones {
		neighbors, err := GetMacrozoneNeighbors(ctx, m.Name, radius)
		if err != nil {
			return nil, err
		}
		for _, n := range neighbors {
			neighborsMap[m.Name] = append(neighborsMap[m.Name], n)
		}
	}

	return neighborsMap, nil
}
