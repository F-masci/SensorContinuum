package health

import "SensorContinuum/internal/edge-hub/comunication"

func isHealthy() bool {
	return comunication.IsConnected()
}
