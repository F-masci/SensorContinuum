package health

import (
	"SensorContinuum/configs/timeouts"
	"SensorContinuum/internal/sensor-agent/comunication"
	"time"
)

var lastValueTimestamp time.Time

// UpdateLastValueTimestamp aggiorna il timestamp dell'ultimo valore ricevuto
func UpdateLastValueTimestamp() {
	lastValueTimestamp = time.Now()
}

// IsHealthy verifica se l'ultimo valore ricevuto è entro il timeout di salute
func isHealthy() bool {
	return time.Since(lastValueTimestamp) < timeouts.IsAliveSensorTimeout && comunication.IsConnected()
}
