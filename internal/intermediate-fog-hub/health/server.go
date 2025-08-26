package health

import (
	"SensorContinuum/pkg/logger"
	"net/http"
	"os"
)

func StartHealthCheckServer(addr string) error {
	http.HandleFunc("/healthz", HealthzHandler)
	return http.ListenAndServe(addr, nil)
}

func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	if isHealthy() {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("ok"))
		if err != nil {
			logger.Log.Error("Failed to write response:", err.Error())
		}
	} else {
		logger.Log.Error("Health check failed")
		http.Error(w, "unhealthy", http.StatusServiceUnavailable)
		os.Exit(1)
	}
}
