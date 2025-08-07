package building

import (
	"SensorContinuum/internal/api-backend/comunication"
	"SensorContinuum/pkg/structure"
	"context"
)

func GetBuildingsList(ctx context.Context, regionName string) ([]structure.Building, error) {
	db, err := comunication.GetRegionPostgresDB(ctx, regionName)
	if err != nil {
		return nil, err
	}
	rows, err := db.Conn().Query(ctx, `
		SELECT id, name, 
			ST_Y(location) as lat, ST_X(location) as lon,
			registration_time, last_comunication
		FROM buildings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buildings []structure.Building
	for rows.Next() {
		var b structure.Building
		if err := rows.Scan(&b.Id, &b.Name, &b.Lat, &b.Lon, &b.RegistrationTime, &b.LastComunication); err != nil {
			return nil, err
		}
		buildings = append(buildings, b)
	}
	return buildings, nil
}

func GetBuildingByID(ctx context.Context, regionName string, id int) (*structure.Building, error) {
	db, err := comunication.GetRegionPostgresDB(ctx, regionName)
	if err != nil {
		return nil, err
	}
	var b structure.Building
	err = db.Conn().QueryRow(ctx, `
		SELECT id, name, 
			ST_Y(location) as lat, ST_X(location) as lon,
			registration_time, last_comunication
		FROM buildings WHERE id = $1`, id).Scan(&b.Id, &b.Name, &b.Lat, &b.Lon, &b.RegistrationTime, &b.LastComunication)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func GetBuildingByName(ctx context.Context, regionName string, name string) (*structure.Building, error) {
	db, err := comunication.GetRegionPostgresDB(ctx, regionName)
	if err != nil {
		return nil, err
	}
	var b structure.Building
	err = db.Conn().QueryRow(ctx, `
		SELECT id, name, 
			ST_Y(location) as lat, ST_X(location) as lon,
			registration_time, last_comunication
		FROM buildings WHERE name = $1`, name).Scan(&b.Id, &b.Name, &b.Lat, &b.Lon, &b.RegistrationTime, &b.LastComunication)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
