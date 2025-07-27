package region

import (
	"SensorContinuum/internal/api-backend/comunication"
	"SensorContinuum/pkg/structure"
	"context"
)

func GetAllRegions(ctx context.Context) ([]structure.Region, error) {
	db, err := comunication.GetPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := db.Conn().Query(ctx, `
		SELECT r.id, r.name, 
			(SELECT COUNT(*) FROM buildings b WHERE b.region_id = r.id) AS building_count
		FROM regions r`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []structure.Region
	for rows.Next() {
		var r structure.Region
		if err := rows.Scan(&r.Id, &r.Name, &r.BuildingCount); err != nil {
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}

func GetRegionByName(ctx context.Context, name string) (*structure.Region, error) {
	db, err := comunication.GetPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	var r structure.Region
	err = db.Conn().QueryRow(ctx, `
		SELECT r.id, r.name, 
			(SELECT COUNT(*) FROM buildings b WHERE b.region_id = r.id) AS building_count
		FROM regions r WHERE r.name = $1`, name).Scan(&r.Id, &r.Name, &r.BuildingCount)
	if err != nil {
		return nil, err
	}

	// Carica i building associati
	buildings, err := getBuildingsByRegionID(ctx, db, r.Id)
	if err != nil {
		return nil, err
	}
	r.Buildings = buildings

	return &r, nil
}

func GetRegionByID(ctx context.Context, id int) (*structure.Region, error) {
	db, err := comunication.GetPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	var r structure.Region
	err = db.Conn().QueryRow(ctx, `
		SELECT r.id, r.name, 
			(SELECT COUNT(*) FROM buildings b WHERE b.region_id = r.id) AS building_count
		FROM regions r WHERE r.id = $1`, id).Scan(&r.Id, &r.Name, &r.BuildingCount)
	if err != nil {
		return nil, err
	}

	// Carica i building associati
	buildings, err := getBuildingsByRegionID(ctx, db, r.Id)
	if err != nil {
		return nil, err
	}
	r.Buildings = buildings

	return &r, nil
}

// Funzione di supporto per caricare i building di una regione
func getBuildingsByRegionID(ctx context.Context, db *comunication.PostgresDB, regionID int) ([]structure.Building, error) {
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
