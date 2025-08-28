package types

import (
	"math"
	"sort"
)

// VariationResult rappresenta il risultato dell'analisi Year-over-Year (YoY) per una specifica macrozona e tipo di dato.
type VariationResult struct {
	Macrozone string  `json:"macrozone"`
	Type      string  `json:"type"`
	Current   float64 `json:"current"`
	Previous  float64 `json:"previous"`
	DeltaPerc float64 `json:"delta_perc"`
	Timestamp int64   `json:"timestamp"`
}

type MacrozoneAnomaly struct {
	MacrozoneName      string            `json:"macrozone"`
	Type               string            `json:"type"`               // tipo di dato (es. PM10, NO2, etc.)
	Variation          VariationResult   `json:"variation"`          // variazione della macrozona
	NeighborMean       float64           `json:"neighbor_mean"`      // media delle variazioni dei vicini
	NeighborStdDev     float64           `json:"neighbor_std_dev"`   // deviazione standard delle variazioni dei vicini
	NeighbourVariation []VariationResult `json:"neighbor_variation"` // variazioni dei vicini
	AbsError           float64           `json:"abs_error"`
	ZScore             float64           `json:"z_score"`
	Timestamp          int64             `json:"timestamp"`
}

func PearsonCorrelation(x []float64, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	n := float64(len(x))
	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}
	numerator := sumXY - (sumX*sumY)/n
	denominator := math.Sqrt((sumX2 - (sumX*sumX)/n) * (sumY2 - (sumY*sumY)/n))
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// rankify converte un array di float64 in ranghi (1-based)
func rankify(values []float64) []float64 {
	n := len(values)
	ranks := make([]float64, n)

	// copia e ordina
	sorted := make([]float64, n)
	copy(sorted, values)
	sort.Float64s(sorted)

	for i, v := range values {
		// trova la posizione (rank)
		for j, s := range sorted {
			if v == s {
				ranks[i] = float64(j + 1)
				break
			}
		}
	}
	return ranks
}

// Spearman calcola il coefficiente di correlazione di Spearman
func spearmanCorrelation(x []float64, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}
	rankX := rankify(x)
	rankY := rankify(y)
	return PearsonCorrelation(rankX, rankY)
}
