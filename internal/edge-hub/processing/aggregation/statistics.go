package aggregation

import (
	"SensorContinuum/pkg/structure"
	"time"
)

// AverageCurrentMinute calcola la media dei valori del minuto corrente.
func AverageCurrentMinute(readings []structure.SensorData) float64 {
	if len(readings) == 0 {
		return 0
	}
	now := time.Now()
	var sum float64
	var count int
	for _, d := range readings {
		t, _ := time.Parse(time.RFC3339, d.Timestamp)
		if t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day() &&
			t.Hour() == now.Hour() && t.Minute() == now.Minute() {
			sum += d.Data
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}
