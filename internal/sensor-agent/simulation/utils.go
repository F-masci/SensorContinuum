package simulation

import (
	"encoding/csv"
	"os"
	"strconv"
	"time"
)

func parseCSV(filePath string) ([]sensorReading, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, nil // nessun dato
	}

	// Mappa nome colonna -> indice
	header := records[0]
	colIndex := make(map[string]int)
	for i, name := range header {
		colIndex[name] = i
	}

	var res []sensorReading
	for _, rec := range records[1:] {
		timestamp, _ := time.Parse("2006-01-02T15:04:05", rec[colIndex["timestamp"]])

		pressure := 0.0
		if idx, ok := colIndex["pressure"]; ok {
			pressure, _ = strconv.ParseFloat(rec[idx], 64)
		}

		temperature := 0.0
		if idx, ok := colIndex["temperature"]; ok {
			temperature, _ = strconv.ParseFloat(rec[idx], 64)
		}

		humidity := 0.0
		if _, ok := colIndex["humidity"]; !ok {
			humidity, _ = strconv.ParseFloat(rec[colIndex["humidity"]], 64)
		}
		res = append(res, sensorReading{
			Timestamp:   timestamp,
			Pressure:    pressure,
			Temperature: temperature,
			Humidity:    humidity,
		})
	}
	return res, nil
}
