package region

import (
	"SensorContinuum/internal/api-backend/building"
	"SensorContinuum/internal/api-backend/comunication"
	"SensorContinuum/pkg/structure"
	"context"
)

func GetAllRegions(ctx context.Context) ([]structure.Region, error) {
	db, err := comunication.GetCloudPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := db.Conn().Query(ctx, `SELECT r.id, r.name FROM regions r`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []structure.Region
	for rows.Next() {
		var r structure.Region
		if err := rows.Scan(&r.Id, &r.Name); err != nil {
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}

func GetRegionByName(ctx context.Context, name string) (*structure.Region, error) {
	db, err := comunication.GetCloudPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	var r structure.Region
	err = db.Conn().QueryRow(ctx, `SELECT r.id, r.name FROM regions r WHERE r.name = $1`, name).Scan(&r.Id, &r.Name)
	if err != nil {
		return nil, err
	}

	// Carica i building associati
	buildings, err := building.GetBuildingsList(ctx, r.Name)
	if err != nil {
		return nil, err
	}
	r.Buildings = buildings
	r.BuildingCount = len(r.Buildings)

	return &r, nil
}

func GetRegionByID(ctx context.Context, id int) (*structure.Region, error) {
	db, err := comunication.GetCloudPostgresDB(ctx)
	if err != nil {
		return nil, err
	}
	var r structure.Region
	err = db.Conn().QueryRow(ctx, `SELECT r.id, r.name FROM regions r WHERE r.id = $1`, id).Scan(&r.Id, &r.Name)
	if err != nil {
		return nil, err
	}

	// Carica i building associati
	buildings, err := building.GetBuildingsList(ctx, r.Name)
	if err != nil {
		return nil, err
	}
	r.Buildings = buildings
	r.BuildingCount = len(r.Buildings)

	return &r, nil
}
