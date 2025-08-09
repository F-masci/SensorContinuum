package aggregation

import (
	"SensorContinuum/pkg/types"
	"time"
)

// AverageInMinute calcola la media dei valori del minuto corrente.
func AverageInMinute(readings []types.SensorData, minute time.Time) float64 {
	if len(readings) == 0 {
		return 0
	}
	var sum float64
	var count int
	for _, d := range readings {
		t, _ := time.Parse(time.RFC3339, d.Timestamp)
		if t.Year() == minute.Year() && t.Month() == minute.Month() && t.Day() == minute.Day() &&
			t.Hour() == minute.Hour() && t.Minute() == minute.Minute() {
			sum += d.Data
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}
