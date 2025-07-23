package filtering

import (
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"math"
)

// IsOutlier controlla se un dato è un outlier basandosi sulla storia recente.
func IsOutlier(data structure.SensorData, historyReadings []structure.SensorData) bool {

	// Se i valori superano dei trashold, li consideriamo outlier.
	if data.Data >= environment.FilteringMaxTreshold || data.Data <= environment.FilteringMinTreshold {
		logger.Log.Debug("Data for sensor ", data.SensorID, "is out of bounds. Value:", data.Data)
		return true
	}

	// Se non abbiamo abbastanza dati, non possiamo fare un calcolo significativo.
	if len(historyReadings) < environment.FilteringMinSamples {
		logger.Log.Debug("Not enough data to calculate outliers for sensor ", data.SensorID, ". Current count:", len(historyReadings))
		return false
	}

	// 1. Calcola la somma e la somma dei quadrati per media e varianza
	var sum, sumSq float64
	for _, reading := range historyReadings {
		sum += reading.Data
		sumSq += reading.Data * reading.Data
	}
	n := float64(len(historyReadings))

	// 2. Calcola media e deviazione standard
	mean := sum / n
	// Varianza = E[X^2] - (E[X])^2
	variance := (sumSq / n) - (mean * mean)
	// Se la varianza è negativa (possibile per errori di floating point), la consideriamo zero.
	if variance < 0 {
		variance = 0
	}
	stdDev := math.Sqrt(variance)

	// 3. Calcola i limiti di accettazione
	lowerBound := mean - environment.FilteringStdDevFactor*stdDev
	upperBound := mean + environment.FilteringStdDevFactor*stdDev

	logger.Log.Info("Outlier check for sensor ", data.SensorID)
	logger.Log.Debug(" - Courrent value: ", data.Data)
	logger.Log.Debug(" - Mean: ", mean)
	logger.Log.Debug(" - StdDev: ", stdDev, " => ", environment.FilteringStdDevFactor, " * StdDev = ", environment.FilteringStdDevFactor*stdDev)
	logger.Log.Debug(" - LowerBound: ", lowerBound)
	logger.Log.Debug(" - UpperBound: ", upperBound)

	// 4. Controlla se il nuovo dato è fuori dai limiti
	return data.Data < lowerBound || data.Data > upperBound

}
