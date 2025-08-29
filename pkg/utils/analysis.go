package utils

import (
	"math"
)

// MovingAverage calcola la media mobile su una finestra
func MovingAverage(series []float64, window int) []float64 {
	if window <= 1 || window > len(series) {
		return series
	}
	smoothed := make([]float64, len(series)-window+1)
	for i := 0; i <= len(series)-window; i++ {
		var sum float64
		for j := 0; j < window; j++ {
			sum += series[i+j]
		}
		smoothed[i] = sum / float64(window)
	}
	return smoothed
}

// Haversine Calcola la distanza tra due coordinate
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // raggio della Terra in km
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// PearsonCorrelation calcola la correlazione di Pearson tra due serie
func PearsonCorrelation(x, y []float64) float64 {
	n := len(x)
	if n != len(y) || n == 0 {
		return math.NaN()
	}
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}
	num := float64(n)*sumXY - sumX*sumY
	den := math.Sqrt((float64(n)*sumX2 - sumX*sumX) * (float64(n)*sumY2 - sumY*sumY))
	if den == 0 {
		return 0
	}
	return num / den
}

// LinearRegressionSlope calcola la pendenza di una retta y = a + bx
func LinearRegressionSlope(y []float64) float64 {
	n := len(y)
	if n == 0 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2 float64
	for i := 0; i < n; i++ {
		x := float64(i)
		sumX += x
		sumY += y[i]
		sumXY += x * y[i]
		sumX2 += x * x
	}
	num := float64(n)*sumXY - sumX*sumY
	den := float64(n)*sumX2 - sumX*sumX
	if den == 0 {
		return 0
	}
	return num / den
}

// MeanRelativeDifference calcola la divergenza media tra due serie
func MeanRelativeDifference(a, b []float64) float64 {
	n := len(a)
	if n != len(b) || n == 0 {
		return math.NaN()
	}
	var sum float64
	for i := 0; i < n; i++ {
		if b[i] != 0 {
			sum += math.Abs(a[i]-b[i]) / math.Abs(b[i])
		}
	}
	return sum / float64(n)
}
