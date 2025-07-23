package simulation

import (
	"SensorContinuum/internal/sensor-agent/environment"
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
	r.Comma = environment.SimulationSeparator
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
		timestamp, _ := time.Parse(environment.SimulationTimestampFormat, rec[colIndex[environment.SimulationTimestampColumn]])

		value := 0.0
		if idx, ok := colIndex[environment.SimulationValueColumn]; ok {
			if value, err = strconv.ParseFloat(rec[idx], 64); err == nil {
				res = append(res, sensorReading{
					Timestamp: timestamp,
					Value:     value,
				})
			}
		}
	}
	return res, nil
}
