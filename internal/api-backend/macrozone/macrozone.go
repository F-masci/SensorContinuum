package macrozone

import (
	"SensorContinuum/internal/api-backend/storage"
	"SensorContinuum/internal/api-backend/zone"
	"SensorContinuum/pkg/types"
	"context"
	"database/sql"
	"errors"
	"math"
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
			ST_Y(ST_Centroid(m.location)) AS lat,
    		ST_X(ST_Centroid(m.location)) AS lon,
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
			ST_Y(ST_Centroid(m.location)) AS lat,
    		ST_X(ST_Centroid(m.location)) AS lon,
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

// GetAggregatedSensorDataByLocation Restituisce i dati aggregati delle macrozone vicine a una posizione
// utilizzando l'ultimo valore aggregato per ogni tipo di sensore
func GetAggregatedSensorDataByLocation(ctx context.Context, lat, lon, radius float64) (*[]types.AggregatedStats, error) {
	// 1. Leggere tutte le macrozone vicine
	macrozones, err := GetMacrozoneByLocation(ctx, lat, lon, radius)
	if err != nil {
		return nil, err
	}

	// 2. Raggruppa le macrozone per regione
	regionsMap := make(map[string][]string)
	for _, m := range macrozones {
		// Aggiungi la macrozona alla lista della regione
		regionsMap[m.RegionName] = append(regionsMap[m.RegionName], m.Name)
	}

	cutoffTime := time.Now().UTC().Add(-2 * time.Hour)

	// 3. Per ogni regione, apri una sola connessione e prendi le aggregazioni di tutte le macrozone di quella regione
	aggregatedMap := make(map[string]*types.AggregatedStats)

	// Per ogni regione, apri una connessione e prendi gli ultimi aggregati di tutte le macrozone
	for region, macrozones := range regionsMap {
		sensorDb, err := storage.GetSensorPostgresDB(ctx, region)
		if err != nil {
			return nil, err
		}

		aggQuery := `
		WITH latest AS (
			SELECT DISTINCT ON (macrozone_name, type)
				macrozone_name,
				type,
				avg_sum,
				avg_count,
				min_value,
				max_value,
				time
			FROM macrozone_aggregated_statistics
			WHERE macrozone_name = ANY($1)
			ORDER BY macrozone_name, type, time DESC
		)
		SELECT *
		FROM latest
	`

		aggRows, err := sensorDb.Conn().Query(ctx, aggQuery, macrozones)
		if err != nil {
			return nil, err
		}
		defer aggRows.Close()

		for aggRows.Next() {
			var macrozoneName, t string
			var sum float64
			var count int
			var minVal, maxVal float64
			var ts time.Time

			if err := aggRows.Scan(&macrozoneName, &t, &sum, &count, &minVal, &maxVal, &ts); err != nil {
				return nil, err
			}

			// Ignora valori troppo vecchi
			if ts.Before(cutoffTime) {
				continue
			}

			// Se tipo non presente, inizializza
			if _, ok := aggregatedMap[t]; !ok {
				aggregatedMap[t] = &types.AggregatedStats{
					Min:   minVal,
					Max:   maxVal,
					Sum:   0,
					Count: 0,
				}
			}

			agg := aggregatedMap[t]

			// Aggiorna somma e count per calcolare la media
			agg.Sum += sum
			agg.Count += count

			// Aggiorna min e max globali
			if minVal < agg.Min {
				agg.Min = minVal
			}
			if maxVal > agg.Max {
				agg.Max = maxVal
			}
		}
	}

	// Trasforma la mappa in slice finale
	results := make([]types.AggregatedStats, 0, len(aggregatedMap))
	for t, agg := range aggregatedMap {
		if agg.Count == 0 {
			continue
		}
		results = append(results, types.AggregatedStats{
			Type:      t,
			Count:     agg.Count,
			Sum:       agg.Sum,
			Avg:       agg.Sum / float64(agg.Count), // media calcolata con sum/count
			Min:       agg.Min,
			Max:       agg.Max,
			Timestamp: time.Now().UTC().Unix(),
		})
	}

	return &results, nil

}

// GetMacrozonesYearlyVariation calcola la variazione YoY per tutte le macrozone di una regione
// Restituisce una mappa annidata: macrozona -> tipo sensore -> VariationResult
func GetMacrozonesYearlyVariation(ctx context.Context, macrozones []types.Macrozone) (map[string]map[string]types.VariationResult, error) {
	// Prendo le variazioni di ieri, di cui ho tutti i dati aggregati
	now := time.Now().UTC()
	dateCurrent := now.Add(-24 * time.Hour)
	datePrev := dateCurrent.AddDate(-1, 0, 0)

	// Raggruppa macrozone per regione
	regionsMap := make(map[string][]string)
	for _, m := range macrozones {
		regionsMap[m.RegionName] = append(regionsMap[m.RegionName], m.Name)
	}

	results := make(map[string]map[string]types.VariationResult)

	for region, macrozoneNames := range regionsMap {
		sensorDb, err := storage.GetSensorPostgresDB(ctx, region)
		if err != nil {
			return nil, err
		}

		aggQuery := `
		SELECT macrozone_name, type, avg_value
		FROM macrozone_daily_agg
		WHERE macrozone_name = ANY($1)
		  AND DATE(day) = DATE($2)
		`

		type stats struct {
			Current float64
			Prev    float64
		}
		data := make(map[string]map[string]*stats)

		// 1. valori attuali
		rows, err := sensorDb.Conn().Query(ctx, aggQuery, macrozoneNames, dateCurrent)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var mz, t string
			var avg float64
			if err := rows.Scan(&mz, &t, &avg); err != nil {
				return nil, err
			}
			if _, ok := data[mz]; !ok {
				data[mz] = make(map[string]*stats)
			}
			if _, ok := data[mz][t]; !ok {
				data[mz][t] = &stats{}
			}
			data[mz][t].Current = avg
		}
		rows.Close()

		// 2. valori anno precedente
		rowsPrev, err := sensorDb.Conn().Query(ctx, aggQuery, macrozoneNames, datePrev)
		if err != nil {
			return nil, err
		}
		for rowsPrev.Next() {
			var mz, t string
			var avg float64
			if err := rowsPrev.Scan(&mz, &t, &avg); err != nil {
				return nil, err
			}
			if _, ok := data[mz]; !ok {
				data[mz] = make(map[string]*stats)
			}
			if _, ok := data[mz][t]; !ok {
				data[mz][t] = &stats{}
			}
			data[mz][t].Prev = avg
		}
		rowsPrev.Close()

		// 3. Calcola variazione percentuale per ciascun sensore
		for mz, sensors := range data {
			for t, s := range sensors {
				if s.Prev == 0 {
					continue
				}
				perc := (s.Current - s.Prev) / s.Prev * 100
				if _, ok := results[mz]; !ok {
					results[mz] = make(map[string]types.VariationResult)
				}
				results[mz][t] = types.VariationResult{
					Macrozone: mz,
					Type:      t,
					Current:   s.Current,
					Previous:  s.Prev,
					DeltaPerc: perc,
					Timestamp: now.Unix(),
				}
			}
		}
	}

	return results, nil
}

// CalculateAnomaliesPerMacrozones calcola, per ogni macrozona e per ogni tipo di sensore,
// la variazione annuale e l'anomalia rispetto ai vicini
// Restituisce una mappa: macrozona -> tipo sensore -> MacrozoneAnomaly
func CalculateAnomaliesPerMacrozones(ctx context.Context, macrozones []types.Macrozone, neighborRadius float64) (map[string]map[string]types.MacrozoneAnomaly, error) {

	// Step 1: calcola mappa vicini
	neighborsMap, err := GetAllMacrozoneNeighbors(ctx, macrozones, neighborRadius)
	if err != nil {
		return nil, err
	}

	// Step 2: crea l’elenco completo di macrozone da cui servono le variazioni
	allNamesMap := make(map[string]types.Macrozone)
	for _, m := range macrozones {
		allNamesMap[m.Name] = m
		for _, n := range neighborsMap[m.Name] {
			if _, ok := allNamesMap[n.Name]; !ok {
				allNamesMap[n.Name] = n
			}
		}
	}
	// Converti in slice per GetMacrozonesYearlyVariation
	allMacrozones := make([]types.Macrozone, 0, len(allNamesMap))
	for _, mz := range allNamesMap {
		allMacrozones = append(allMacrozones, mz)
	}

	// Step 3: chiedi le variazioni per tutte
	// FIXME: prendere i dati dal DB, piuttosto che ricalcolarli ogni volta
	macrozoneVariations, err := GetMacrozonesYearlyVariation(ctx, allMacrozones)
	if err != nil {
		return nil, err
	}

	// Step 4: calcola anomalie per ciascun tipo sensore
	anomalies := make(map[string]map[string]types.MacrozoneAnomaly)

	for _, m := range macrozones {
		myVars, ok := macrozoneVariations[m.Name]
		if !ok {
			continue
		}

		neighbors := neighborsMap[m.Name]
		if len(neighbors) == 0 {
			continue
		}

		for sensorType, myVariation := range myVars {
			neighbourVariation := make([]types.VariationResult, 0)

			// raccolta delle variazioni dei vicini
			neighborDeltas := make([]float64, 0)
			for _, n := range neighbors {
				if nv, ok := macrozoneVariations[n.Name]; ok {
					if v, ok2 := nv[sensorType]; ok2 {
						neighbourVariation = append(neighbourVariation, v)
						neighborDeltas = append(neighborDeltas, v.DeltaPerc)
					}
				}
			}

			count := len(neighborDeltas)
			if count == 0 {
				continue
			}

			// calcola media
			var sum float64
			for _, d := range neighborDeltas {
				sum += d
			}
			mean := sum / float64(count)

			// calcola deviazione standard
			var variance float64
			for _, d := range neighborDeltas {
				variance += (d - mean) * (d - mean)
			}
			stdDev := math.Sqrt(variance / float64(count))

			// calcola z-score
			var zScore float64
			if stdDev != 0 {
				zScore = (myVariation.DeltaPerc - mean) / stdDev
			} else {
				zScore = 0 // tutti i vicini hanno lo stesso valore
			}

			if _, ok := anomalies[m.Name]; !ok {
				anomalies[m.Name] = make(map[string]types.MacrozoneAnomaly)
			}
			anomalies[m.Name][sensorType] = types.MacrozoneAnomaly{
				MacrozoneName:      m.Name,
				Type:               sensorType,
				Variation:          myVariation,
				NeighborMean:       mean,
				NeighborStdDev:     stdDev,
				NeighbourVariation: neighbourVariation,
				AbsError:           math.Abs(myVariation.DeltaPerc - mean),
				ZScore:             zScore,
				Timestamp:          myVariation.Timestamp,
			}
		}
	}

	return anomalies, nil
}
