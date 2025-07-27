package building

import (
	"SensorContinuum/internal/api-backend/comunication"
	"SensorContinuum/pkg/structure"
	"context"
)

func GetAllBuildings(ctx context.Context) ([]structure.Building, error) {
	db, err := comunication.GetPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := db.Conn().Query(ctx, `
		SELECT id, region_id, name, 
			ST_Y(location) as lat, ST_X(location) as lon
		FROM buildings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buildings []structure.Building
	for rows.Next() {
		var b structure.Building
		if err := rows.Scan(&b.Id, &b.RegionId, &b.Name, &b.Lat, &b.Lon); err != nil {
			return nil, err
		}
		buildings = append(buildings, b)
	}
	return buildings, nil
}

func GetBuildingByID(ctx context.Context, id int) (*structure.Building, error) {
	db, err := comunication.GetPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	var b structure.Building
	err = db.Conn().QueryRow(ctx, `
		SELECT id, region_id, name, 
			ST_Y(location) as lat, ST_X(location) as lon
		FROM buildings WHERE id = $1`, id).Scan(&b.Id, &b.RegionId, &b.Name, &b.Lat, &b.Lon)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func GetBuildingByName(ctx context.Context, name string) (*structure.Building, error) {
	db, err := comunication.GetPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	var b structure.Building
	err = db.Conn().QueryRow(ctx, `
		SELECT id, region_id, name, 
			ST_Y(location) as lat, ST_X(location) as lon
		FROM buildings WHERE name = $1`, name).Scan(&b.Id, &b.RegionId, &b.Name, &b.Lat, &b.Lon)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func GetBuildingsByRegionID(ctx context.Context, regionID int) ([]structure.Building, error) {
	db, err := comunication.GetPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := db.Conn().Query(ctx, `
		SELECT id, region_id, name, 
			ST_Y(location) as lat, ST_X(location) as lon
		FROM buildings WHERE region_id = $1`, regionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buildings []structure.Building
	for rows.Next() {
		var b structure.Building
		if err := rows.Scan(&b.Id, &b.RegionId, &b.Name, &b.Lat, &b.Lon); err != nil {
			return nil, err
		}
		buildings = append(buildings, b)
	}
	return buildings, nil
}
