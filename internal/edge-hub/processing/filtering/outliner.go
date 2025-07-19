package filtering

import (
	_ "SensorContinuum/internal/edge-hub/processing"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"math"
)

const (
	// MinSamplesForCalculation è il numero minimo di dati necessari prima di iniziare a calcolare gli outlier.
	MinSamplesForCalculation = 5
	// StdDevFactor determina quanto un valore si deve discostare dalla media per essere un outlier.
	// Un valore comune è tra 2 e 3. Con 2, circa il 95% dei dati (in una distribuzione normale) è considerato valido.
	StdDevFactor = 2.0
)

// IsOutlier controlla se un dato è un outlier basandosi sulla storia recente.
func IsOutlier(data structure.SensorData, historyReadings []structure.SensorData) bool {
	// Se non abbiamo abbastanza dati, non possiamo fare un calcolo significativo.
	if len(historyReadings) < MinSamplesForCalculation {
		logger.Log.Debug("Not enough data to calculate outliers for sensor", data.SensorID, ". Current count:", len(historyReadings))
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
	lowerBound := mean - StdDevFactor*stdDev
	upperBound := mean + StdDevFactor*stdDev

	logger.Log.Debug("Outlier check for sensor ", data.SensorID)
	logger.Log.Info("courrent value: ", data.Data)
	logger.Log.Info("mean: ", mean)
	logger.Log.Info("stdDev: ", stdDev)
	logger.Log.Info("lowerBound: ", lowerBound)
	logger.Log.Info("upperBound: ", upperBound)

	// 4. Controlla se il nuovo dato è fuori dai limiti
	if data.Data < lowerBound || data.Data > upperBound {
		return true
	}

	return false
}
