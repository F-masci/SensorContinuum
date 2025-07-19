package simulation

import (
	"SensorContinuum/configs/simulation"
	"SensorContinuum/pkg/logger"
	"math"
	"math/rand"
	"time"
)

var (
	readings    []sensorReading
	statsByHour map[int]distribution
)

type distribution struct {
	Mean float64
	Std  float64
}

// SensorReading rappresenta una riga del CSV
type sensorReading struct {
	Timestamp   time.Time
	Pressure    float64
	Temperature float64
	Humidity    float64
}

// setupDistribution carica il file e prepara la distribuzione
func setupDistribution(filePath string) error {
	var err error

	readings, err = parseCSV(filePath)
	if err != nil {
		return err
	}

	computeStatsByHour()

	return nil
}

// generateValue genera una lettura randomica basata sulla distribuzione
func generateValue() sensorReading {
	now := time.Now()
	return generateRandomReading(now)
}

// computeStatsByHour calcola la media e la deviazione standard per le letture di una specifica ora
// Filtra le letture per l'ora richiesta, calcola la media e la deviazione standard, utili per generare valori randomici realistici.
func computeStatsByHour() {
	statsByHour = make(map[int]distribution)
	for hour := 0; hour < 24; hour++ {
		var vals []float64
		for _, r := range readings {
			if r.Timestamp.Hour() == hour {
				vals = append(vals, r.Temperature)
			}
		}
		if len(vals) == 0 {
			statsByHour[hour] = distribution{0, 0}
			continue
		}
		var sum float64
		for _, v := range vals {
			sum += v
		}
		mean := sum / float64(len(vals))
		var std float64
		for _, v := range vals {
			std += (v - mean) * (v - mean)
		}
		std = std / float64(len(vals))
		std = math.Sqrt(std)
		statsByHour[hour] = distribution{mean, std}
	}
}

// generateRandomReading genera una lettura randomica basata sulla distribuzione
func generateRandomReading(datetime time.Time) sensorReading {
	stats := statsByHour[datetime.Hour()]
	logger.Log.Debug("Generating random reading for hour: ", datetime.Hour(), " with mean: ", stats.Mean, " and std: ", stats.Std)

	// Genera un valore casuale basato sulla distribuzione normale
	temp := rand.NormFloat64() * stats.Std

	// Aggiunge un outlier con una probabilità definita
	if rand.Float64() < simulation.OUTLIER_PROBABILITY {
		logger.Log.Debug("Generating outlier")
		temp *= simulation.OUTLIER_MULTIPLIER // Moltiplica per il moltiplicatore per generare un outlier
		temp += simulation.OUTLIER_ADDITION   // Aggiunge un valore per aumentare il centro dell'outlier
	}

	// Aggiunge la media per centrare il valore
	temp += stats.Mean

	return sensorReading{
		Timestamp:   datetime,
		Temperature: temp,
	}
}
