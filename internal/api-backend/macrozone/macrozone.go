package macrozone

import (
	"SensorContinuum/internal/api-backend/environment"
	"SensorContinuum/internal/api-backend/storage"
	"SensorContinuum/internal/api-backend/zone"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"SensorContinuum/pkg/utils"
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

// GetAggregatedDataByLocation Restituisce i dati aggregati delle macrozone vicine a una posizione
// utilizzando l'ultimo valore aggregato per ogni tipo di sensore
func GetAggregatedDataByLocation(ctx context.Context, lat, lon, radius float64) ([]types.AggregatedStats, error) {
	// 1. Leggere tutte le macrozone vicine
	nearestMacrozones, err := GetMacrozonesByLocation(ctx, lat, lon, radius)
	if err != nil {
		return nil, err
	}

	for _, macrozone := range nearestMacrozones {
		logger.Log.Info("Found macrozone:", macrozone.Name, " at ", macrozone.Lat, macrozone.Lon, " in region ", macrozone.RegionName, " distance ", math.Round(utils.Haversine(lat, lon, macrozone.Lat, macrozone.Lon)))
	}

	if len(nearestMacrozones) == 0 {
		return nil, nil
	}

	// 2. Raggruppa le macrozone per regione
	regionsMap := make(map[string][]string)
	for _, m := range nearestMacrozones {
		// Aggiungi la macrozona alla lista della regione
		regionsMap[m.RegionName] = append(regionsMap[m.RegionName], m.Name)
	}

	cutoffTime := time.Now().UTC().Add(-environment.AggregatedDataCutOff).Truncate(environment.AggregatedDataCutOff)

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
				avg_value,
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
			var avg, sum float64
			var count int
			var minVal, maxVal float64
			var ts time.Time

			if err := aggRows.Scan(&macrozoneName, &t, &avg, &sum, &count, &minVal, &maxVal, &ts); err != nil {
				return nil, err
			}

			// Ignora valori troppo vecchi
			if ts.Before(cutoffTime) {
				continue
			}

			// Se tipo non presente, inizializza
			if _, ok := aggregatedMap[t]; !ok {
				aggregatedMap[t] = &types.AggregatedStats{
					Type:          t,
					Min:           minVal,
					Max:           maxVal,
					Sum:           0,
					Count:         0,
					WeightedSum:   0,
					WeightedCount: 0,
					Timestamp:     time.Now().UTC().Unix(),
				}
			}

			agg := aggregatedMap[t]

			// Trova la macrozona corrispondente per lat/lon
			var mz types.Macrozone
			for _, m := range nearestMacrozones {
				if m.Name == macrozoneName {
					mz = m
					break
				}
			}

			dist := utils.Haversine(lat, lon, mz.Lat, mz.Lon)
			if dist == 0 {
				dist = 0.001 // evita divisione per zero
			}

			// Aggiorna somma, count e weight per calcolare la media
			weight := 1 / dist
			agg.WeightedSum += avg * weight
			agg.WeightedCount += weight
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
	for _, agg := range aggregatedMap {
		if agg.Count == 0 || agg.WeightedCount == 0 {
			continue
		}
		// Calcola la media ponderata
		agg.Avg = agg.Sum / float64(agg.Count)
		agg.WeightedAvg = agg.WeightedSum / float64(agg.WeightedCount)
		results = append(results, *agg)
	}

	return results, nil

}

// GetMacrozonesYearlyVariation Cerca le variazioni YoY per tutte le macrozone.
// Se i dati non sono presenti nel DB (ultimi 2 giorni), li calcola al momento.
// Restituisce una mappa annidata: macrozona -> tipo sensore -> VariationResult
func GetMacrozonesYearlyVariation(ctx context.Context, macrozones []types.Macrozone, day time.Time) (map[string]map[string]types.VariationResult, error) {

	// Controlla che ci siano macrozone
	sensorDb, err := storage.GetSensorPostgresDB(ctx, macrozones[0].RegionName)
	if err != nil {
		return nil, err
	}

	// Prova a leggere le variazioni dal DB
	rows, err := sensorDb.Conn().Query(ctx, `
		SELECT macrozone, type, current, previous, delta_perc, time
		FROM macrozone_yearly_variation
		WHERE DATE(time) = $1
		  AND macrozone = ANY($2)
	`, day.UTC().Format("2006-01-02"), func() []string {
		names := make([]string, 0, len(macrozones))
		for _, m := range macrozones {
			names = append(names, m.Name)
		}
		return names
	}())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]map[string]types.VariationResult)
	found := false

	for rows.Next() {
		found = true
		var mzName, t string
		var current, previous, deltaPerc float64
		var ts time.Time

		if err := rows.Scan(&mzName, &t, &current, &previous, &deltaPerc, &ts); err != nil {
			return nil, err
		}

		if _, ok := results[mzName]; !ok {
			results[mzName] = make(map[string]types.VariationResult)
		}
		results[mzName][t] = types.VariationResult{
			Macrozone: mzName,
			Type:      t,
			Current:   current,
			Previous:  previous,
			DeltaPerc: deltaPerc,
			Timestamp: ts.Unix(),
		}
	}

	// Se non ci sono dati, calcola per quella data
	if !found {
		return ComputeMacrozonesYearlyVariation(ctx, macrozones, day)
	}

	return results, nil

}

// ComputeMacrozonesYearlyVariation calcola la variazione YoY per tutte le macrozone
// Restituisce una mappa annidata: macrozona -> tipo sensore -> VariationResult
func ComputeMacrozonesYearlyVariation(ctx context.Context, macrozones []types.Macrozone, dateCurrent time.Time) (map[string]map[string]types.VariationResult, error) {

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

		// Aggiorna la vista materializzata se necessario
		err = storage.CheckAndRefreshAggregateView(ctx, sensorDb, "macrozone_daily_agg", dateCurrent)
		err = storage.CheckAndRefreshAggregateView(ctx, sensorDb, "macrozone_daily_agg", datePrev)

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

		// 3. Calcola variazione percentuale e popola struttura risultati
		mvBatch := make([]types.VariationResult, 0)
		for mz, sensors := range data {
			for t, s := range sensors {
				if s.Prev == 0 {
					continue
				}
				perc := (s.Current - s.Prev) / s.Prev * 100
				if _, ok := results[mz]; !ok {
					results[mz] = make(map[string]types.VariationResult)
				}
				r := types.VariationResult{
					Macrozone: mz,
					Type:      t,
					Current:   s.Current,
					Previous:  s.Prev,
					DeltaPerc: perc,
					Timestamp: dateCurrent.Unix(),
				}
				results[mz][t] = r
				mvBatch = append(mvBatch, r)
			}
		}

		// Salva i risultati nel database
		err = storage.SaveVariationResults(ctx, sensorDb, mvBatch)
		if err != nil {
			return nil, err
		}

	}

	return results, nil
}

// GetMacrozonesAnomalies restituisce le anomalie per una data specifica.
// Se non sono presenti nel DB, le calcola al momento.
func GetMacrozonesAnomalies(ctx context.Context, macrozones []types.Macrozone, radius float64, date time.Time) (map[string]map[string]types.MacrozoneAnomaly, error) {
	sensorDb, err := storage.GetSensorPostgresDB(ctx, macrozones[0].RegionName)
	if err != nil {
		return nil, err
	}

	// Query unica: cerca anomalie per la data specifica
	rows, err := sensorDb.Conn().Query(ctx, `
		SELECT macrozone, type, current, previous, delta_perc,
		       neighbor_mean, neighbor_std_dev, abs_error, z_score, time
		FROM macrozone_yearly_anomalies
		WHERE DATE(time) = $1
		  AND macrozone = ANY($2)
	`, date.UTC().Format("2006-01-02"), func() []string {
		names := make([]string, 0, len(macrozones))
		for _, m := range macrozones {
			names = append(names, m.Name)
		}
		return names
	}())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]map[string]types.MacrozoneAnomaly)
	found := false

	for rows.Next() {
		found = true
		var mzName, t string
		var current, previous, deltaPerc float64
		var neighborMean, neighborStd, absErr, zScore float64
		var ts time.Time

		if err := rows.Scan(&mzName, &t, &current, &previous, &deltaPerc,
			&neighborMean, &neighborStd, &absErr, &zScore, &ts); err != nil {
			return nil, err
		}

		if _, ok := results[mzName]; !ok {
			results[mzName] = make(map[string]types.MacrozoneAnomaly)
		}

		results[mzName][t] = types.MacrozoneAnomaly{
			MacrozoneName: mzName,
			Type:          t,
			Variation: types.VariationResult{
				Macrozone: mzName,
				Type:      t,
				Current:   current,
				Previous:  previous,
				DeltaPerc: deltaPerc,
				Timestamp: ts.Unix(),
			},
			NeighborMean:   neighborMean,
			NeighborStdDev: neighborStd,
			AbsError:       absErr,
			ZScore:         zScore,
			Timestamp:      ts.Unix(),
		}
	}

	// Se non ci sono dati, calcola al volo
	if !found {
		return ComputeMacrozonesAnomalies(ctx, macrozones, radius, date)
	}

	return results, nil
}

// ComputeMacrozonesAnomalies calcola, per ogni macrozona e per ogni tipo di sensore,
// la variazione annuale e l'anomalia rispetto ai vicini
// Restituisce una mappa: macrozona -> tipo sensore -> MacrozoneAnomaly
func ComputeMacrozonesAnomalies(ctx context.Context, macrozones []types.Macrozone, neighborRadius float64, date time.Time) (map[string]map[string]types.MacrozoneAnomaly, error) {

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

	// Step 3: chiedi le variazioni per tutte le macrozone
	macrozoneVariations, err := GetMacrozonesYearlyVariation(ctx, allMacrozones, date)
	if err != nil {
		return nil, err
	}

	// Raggruppa macrozone per regione
	regionsMap := make(map[string][]string)
	for _, m := range macrozones {
		regionsMap[m.RegionName] = append(regionsMap[m.RegionName], m.Name)
	}

	// Step 4: calcola anomalie per ciascun tipo sensore
	anomalies := make(map[string]map[string]types.MacrozoneAnomaly)

	for region, macrozoneName := range regionsMap {

		sensorDb, err := storage.GetSensorPostgresDB(ctx, region)
		if err != nil {
			return nil, err
		}

		maBatch := make([]types.MacrozoneAnomaly, 0)
		for _, m := range macrozoneName {
			macroVars, ok := macrozoneVariations[m]
			if !ok {
				continue
			}

			neighbors := neighborsMap[m]
			if len(neighbors) == 0 {
				continue
			}

			for sensorType, myVariation := range macroVars {
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

				if _, ok := anomalies[m]; !ok {
					anomalies[m] = make(map[string]types.MacrozoneAnomaly)
				}
				a := types.MacrozoneAnomaly{
					MacrozoneName:      m,
					Type:               sensorType,
					Variation:          myVariation,
					NeighborMean:       mean,
					NeighborStdDev:     stdDev,
					NeighbourVariation: neighbourVariation,
					AbsError:           math.Abs(myVariation.DeltaPerc - mean),
					ZScore:             zScore,
					Timestamp:          myVariation.Timestamp,
				}
				anomalies[m][sensorType] = a
				maBatch = append(maBatch, a)
			}

		}

		// Salva i risultati nel database
		err = storage.SaveAnomaliesResults(ctx, sensorDb, maBatch)
		if err != nil {
			return nil, err
		}

	}

	return anomalies, nil
}

// GetMacrozonesTrendsSimilarity Restituisce la similarità dei trend
// di ciascuna macrozona rispetto alla sua regione, considerando tutti i tipi di rilevazione.
// Se i dati non sono presenti nel DB, li calcola al momento.
// Restituisce una mappa annidata: macrozona -> tipo di rilevazione -> TrendSimilarityResult
func GetMacrozonesTrendsSimilarity(ctx context.Context, macrozones []types.Macrozone, days int, date time.Time) (map[string]map[string]types.TrendSimilarityResult, error) {
	sensorDb, err := storage.GetSensorPostgresDB(ctx, macrozones[0].RegionName)
	if err != nil {
		return nil, err
	}

	// Prova a leggere le similarità dal DB
	rows, err := sensorDb.Conn().Query(ctx, `
		SELECT macrozone, type, correlation, slope_macro, slope_region, divergence, time
		FROM macrozone_trends_similarity
		WHERE DATE(time) = $1
		  AND macrozone = ANY($2)
	`, date.UTC().Format("2006-01-02"), func() []string {
		names := make([]string, 0, len(macrozones))
		for _, m := range macrozones {
			names = append(names, m.Name)
		}
		return names
	}())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]map[string]types.TrendSimilarityResult)
	found := false

	for rows.Next() {
		found = true
		var mzName, t string
		var correlation, slopeMacro, slopeRegion, divergence float64
		var ts time.Time

		if err := rows.Scan(&mzName, &t, &correlation, &slopeMacro, &slopeRegion, &divergence, &ts); err != nil {
			return nil, err
		}

		if _, ok := results[mzName]; !ok {
			results[mzName] = make(map[string]types.TrendSimilarityResult)
		}
		results[mzName][t] = types.TrendSimilarityResult{
			MacrozoneName: mzName,
			Type:          t,
			Correlation:   correlation,
			SlopeMacro:    slopeMacro,
			SlopeRegion:   slopeRegion,
			Divergence:    divergence,
			Timestamp:     date.UTC().Unix(),
		}
	}

	// Se non ci sono dati, calcola per quella data
	if !found {
		return ComputeMacrozonesTrendsSimilarity(ctx, macrozones, days, date)
	}

	return results, nil

}

// ComputeMacrozonesTrendsSimilarity calcola la similarità dei trend
// di ciascuna macrozona rispetto alla sua regione, considerando tutti i tipi di rilevazione.
// Restituisce una mappa annidata: macrozona -> tipo di rilevazione -> TrendSimilarityResult
func ComputeMacrozonesTrendsSimilarity(ctx context.Context, macrozones []types.Macrozone, days int, date time.Time) (map[string]map[string]types.TrendSimilarityResult, error) {
	results := make(map[string]map[string]types.TrendSimilarityResult) // mappa finale dei risultati

	// Range di giorni
	startDate := date.AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour) // giorno iniziale
	endDate := date.Truncate(24 * time.Hour)                              // giorno finale

	// 1. Raggruppa le macrozone per regione per ridurre query al DB
	regionsMap := make(map[string][]types.Macrozone)
	for _, m := range macrozones {
		regionsMap[m.RegionName] = append(regionsMap[m.RegionName], m)
	}

	// 2. Itera sulle regioni
	for regionName, mzList := range regionsMap {
		sensorDb, err := storage.GetSensorPostgresDB(ctx, regionName)
		if err != nil {
			return nil, err
		}

		tsBatch := make([]types.TrendSimilarityResult, 0)

		// Aggiorna la view solo per il range specificato
		for _, viewName := range []string{"region_daily_agg", "macrozone_daily_agg"} {
			for d := 0; d < days; d++ {
				aggDate := startDate.AddDate(0, 0, d)
				err = storage.CheckAndRefreshAggregateView(ctx, sensorDb, viewName, aggDate)
				if err != nil {
					return nil, err
				}
			}
		}

		// --- Recupera dati della regione ---
		queryRegion := `
            SELECT type, day, avg_value, min_value, max_value, total_sum, total_count
            FROM region_daily_agg
            WHERE day BETWEEN $1 AND $2
            ORDER BY type, day
        `
		rows, err := sensorDb.Conn().Query(ctx, queryRegion, startDate, endDate)
		if err != nil {
			return nil, err
		}

		// regionData[type][day] = AggregatedStats
		regionData := make(map[string]map[time.Time]*types.AggregatedStats)
		for rows.Next() {
			var t string
			var day time.Time
			var avg, min, max, sum float64
			var count int
			if err := rows.Scan(&t, &day, &avg, &min, &max, &sum, &count); err != nil {
				return nil, err
			}
			if _, ok := regionData[t]; !ok {
				regionData[t] = make(map[time.Time]*types.AggregatedStats)
			}
			regionData[t][day] = &types.AggregatedStats{
				Timestamp: day.Unix(),
				Region:    regionName,
				Type:      t,
				Avg:       avg,
				Min:       min,
				Max:       max,
				Sum:       sum,
				Count:     count,
			}
		}
		rows.Close()

		// --- Recupera tutti i dati delle macrozone della regione ---
		mzNames := make([]string, 0, len(mzList))
		for _, mz := range mzList {
			mzNames = append(mzNames, mz.Name)
		}

		queryMacro := `
            SELECT macrozone_name, type, day, avg_value, min_value, max_value, total_sum, total_count
            FROM macrozone_daily_agg
            WHERE macrozone_name = ANY($1)
              AND day BETWEEN $2 AND $3
            ORDER BY macrozone_name, type, day
        `
		rows, err = sensorDb.Conn().Query(ctx, queryMacro, mzNames, startDate, endDate)
		if err != nil {
			return nil, err
		}

		// macroData[macrozona][type][day] = AggregatedStats
		macroData := make(map[string]map[string]map[time.Time]*types.AggregatedStats)
		for rows.Next() {
			var mzName, t string
			var day time.Time
			var avg, min, max, sum float64
			var count int
			if err := rows.Scan(&mzName, &t, &day, &avg, &min, &max, &sum, &count); err != nil {
				return nil, err
			}
			if _, ok := macroData[mzName]; !ok {
				macroData[mzName] = make(map[string]map[time.Time]*types.AggregatedStats)
			}
			if _, ok := macroData[mzName][t]; !ok {
				macroData[mzName][t] = make(map[time.Time]*types.AggregatedStats)
			}
			macroData[mzName][t][day] = &types.AggregatedStats{
				Timestamp: day.Unix(),
				Macrozone: mzName,
				Region:    regionName,
				Type:      t,
				Avg:       avg,
				Min:       min,
				Max:       max,
				Sum:       sum,
				Count:     count,
			}
		}
		rows.Close()

		// --- Calcola trend similarity per ciascuna macrozona e tipo ---
		for _, mz := range mzList {
			for t, regSeries := range regionData {
				mzSeriesMap, ok := macroData[mz.Name][t]
				if !ok {
					continue // se la macrozona non ha dati per questo tipo, salta
				}

				// Allineamento temporale usando AggregatedStats
				alignedMacro := make([]types.AggregatedStats, 0, days)
				alignedRegion := make([]types.AggregatedStats, 0, days)
				for day, regAgg := range regSeries {
					if mzAgg, ok := mzSeriesMap[day]; ok {
						// Aggiungi solo se entrambi hanno dati per quel giorno
						alignedMacro = append(alignedMacro, *mzAgg)
						alignedRegion = append(alignedRegion, *regAgg)
					}
				}

				// Se la serie è troppo corta, non calcoliamo il trend
				// if len(alignedMacro) < 5 {
				// 	 continue
				// }

				// --- Estrai le medie per calcolo trend ---
				macroSeries := make([]float64, len(alignedMacro))
				regionSeries := make([]float64, len(alignedRegion))
				for i := range alignedMacro {
					macroSeries[i] = alignedMacro[i].Avg
					regionSeries[i] = alignedRegion[i].Avg
				}

				// --- Analisi statistica ---
				trendMz := utils.MovingAverage(macroSeries, 3)      // lisciamento
				trendReg := utils.MovingAverage(regionSeries, 3)    // lisciamento
				corr := utils.PearsonCorrelation(trendMz, trendReg) // correlazione Pearson
				slopeMz := utils.LinearRegressionSlope(trendMz)     // pendenza macrozona
				slopeReg := utils.LinearRegressionSlope(trendReg)   // pendenza regione
				div := utils.MeanRelativeDifference(trendMz, trendReg)

				if math.IsNaN(corr) {
					corr = 0
				}

				if math.IsNaN(slopeMz) {
					slopeMz = 0
				}

				if math.IsNaN(slopeReg) {
					slopeReg = 0
				}

				if math.IsNaN(div) {
					div = 0
				}

				// Inizializza la mappa interna se necessario
				if _, ok := results[mz.Name]; !ok {
					results[mz.Name] = make(map[string]types.TrendSimilarityResult)
				}

				// Salva il risultato
				tsr := types.TrendSimilarityResult{
					MacrozoneName:   mz.Name,
					RegionName:      regionName,
					Type:            t,
					Correlation:     corr,
					SlopeMacro:      slopeMz,
					SlopeRegion:     slopeReg,
					Divergence:      div,
					MacrozoneSeries: alignedMacro,
					RegionSeries:    alignedRegion,
				}
				tsBatch = append(tsBatch, tsr)
				results[mz.Name][t] = tsr
			}
		}

		// Salva i risultati nel database
		err = storage.SaveTrendSimilarityResults(ctx, sensorDb, tsBatch, date)
		if err != nil {
			return nil, err
		}

	}

	return results, nil
}
