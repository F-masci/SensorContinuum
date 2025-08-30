package aggregation

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"time"
)

// Run è la funzione che avvia il processo di aggregazione periodica.
// Essa avvia un ticker che esegue l'aggregazione ogni intervallo di tempo
// definito in environment.AggregationInterval.
// Questa funzione viene eseguita in una goroutine separata.
func Run(ctx context.Context) {

	// Avvio del ticker per l'aggregazione periodica
	statsTicker := time.NewTicker(environment.AggregationInterval)
	logger.Log.Info("Aggregation ticker started, aggregating data every ", environment.AggregationInterval.Minutes(), " minutes from now.")
	defer statsTicker.Stop()

	for {
		select {
		// Permette uno spegnimento pulito quando il contesto viene annullato
		case <-ctx.Done():
			logger.Log.Info("Stopping aggregator...")
			return
		case <-statsTicker.C:
			logger.Log.Info("Execution of aggregation started")
			AggregateSensorData(ctx)
		}
	}

}

// AggregateSensorData è la funzione che viene eseguita per l'aggregazione dei dati dei sensori.
// Questa funzione calcola le statistiche aggregate (min, max, avg, sum, count)
// per ogni tipo di sensore e per ogni zona, nell'intervallo di tempo specificato.
// Le statistiche vengono poi salvate nella tabella di cache per essere inviate al Intermediate Hub.
func AggregateSensorData(ctx context.Context) {

	// 1. Calcola gli intervalli allineati per l'aggregazione
	lastAggregation, err := storage.GetLastMacrozoneAggregatedData(ctx)
	if err != nil {
		logger.Log.Error("Failed to get last aggregation time: ", err)
		return
	}

	var alignedStartTime time.Time
	if lastAggregation != (types.AggregatedStats{}) {
		// Se esiste una precedente aggregazione, usiamo il suo timestamp come inizio del nuovo intervallo
		alignedStartTime = time.Unix(lastAggregation.Timestamp, 0).UTC()
		logger.Log.Info("Starting aggregation data from last aggregation time ", alignedStartTime.Format(time.RFC3339))
	} else {
		// Se non esiste una precedente aggregazione,
		// calcola il tempo di inizio considerando l'offset
		// e allineandolo all'intervallo di aggregazione
		// Ad esempio, se ora sono le 15:48:30 e l'offset è -10min,
		// alignedEndTime sarà 15:38:00 (portato a 15:30:00)
		// e alignedStartTime sarà 15:23:00 (portato a 15:15:00)
		// Questo assicura che le aggregazioni siano sempre allineate a intervalli regolari
		// e non inizino in momenti casuali.
		now := time.Now().UTC()
		alignedStartTime = now.Add(environment.AggregationStartingOffset + environment.AggregationFetchOffset - environment.AggregationInterval).Truncate(environment.AggregationInterval)
		logger.Log.Info("No previous aggregation found, starting aggregation data from ", alignedStartTime.Format(time.RFC3339))
	}

	// Calcola il massimo tempo di fine allineato
	maxAlignedEndTime := time.Now().UTC().Add(environment.AggregationFetchOffset).Truncate(environment.AggregationInterval)
	if !alignedStartTime.Before(maxAlignedEndTime) {
		logger.Log.Warn("Aggregation start time is not before the maximum end time, skipping aggregation")
		return
	}

	// Calcola gli intervalli di aggregazione da elaborare
	var intervals []struct{ Start, End time.Time }
	for start := alignedStartTime; start.Before(maxAlignedEndTime); start = start.Add(environment.AggregationInterval) {
		end := start.Add(environment.AggregationInterval)
		if end.After(maxAlignedEndTime) {
			logger.Log.Error("End time exceeds maximum aligned end time, skipping this interval")
			break
		}
		intervals = append(intervals, struct{ Start, End time.Time }{Start: start, End: end})
	}

	for _, interval := range intervals {
		alignedStartTime = interval.Start
		alignedEndTime := interval.End
		logger.Log.Info("Processing aggregation interval from ", alignedStartTime.Format(time.RFC3339), " to ", alignedEndTime.Format(time.RFC3339))

		// 2. Esegue la query per ottenere le statistiche dei dati arrivati nell'intervallo alignedStartTime e alignedEndTime
		stats, err := storage.GetZoneAggregatedData(ctx, alignedStartTime, alignedEndTime)
		if err != nil {
			logger.Log.Error("Failed to calculate periodic statistics: ", err)
			continue
		}

		if len(stats) == 0 {
			logger.Log.Warn("No data to send, skipping aggregation")
			continue
		}

		// Calcola le statistiche aggregate a livello di macrozona
		macrozoneStats := computeMacrozoneAggregate(stats, alignedEndTime)
		stats = append(stats, macrozoneStats...)

		// 3. Salva le statistiche aggregate nella tabella outbox
		for _, stat := range stats {
			// Arricchiamo la statistica con dati contestuali prima di salvarla
			stat.Timestamp = alignedEndTime.UTC().Unix()
			stat.Macrozone = environment.EdgeMacrozone

			logger.Log.Info("Statistics calculated for the type: ", stat.Type, ", min: ", stat.Min, ", max: ", stat.Max, ", avg: ", stat.Avg)

			if err := storage.InsertAggregatedStats(ctx, stat); err != nil {
				logger.Log.Error("Failure to save statistics to cache for type ", stat.Type, ": ", err)
				// Non ci fermiamo, proviamo a salvare le altre
				continue
			}
			logger.Log.Info("Statistics successfully saved to cache for type: ", stat.Type)
		}
	}
}

// computeMacrozoneAggregate calcola le statistiche aggregate a livello di macrozona
// a partire dalle statistiche aggregate delle varie zone
func computeMacrozoneAggregate(aggregatedStats []types.AggregatedStats, timestamp time.Time) []types.AggregatedStats {

	if len(aggregatedStats) == 0 {
		logger.Log.Warn("No aggregated statistics available to compute macrozone aggregate")
		return nil
	}

	// Creo una mappatura per tenere traccia della somma, del minimo e del massimo per ogni tipo di sensore
	values := make(map[string]types.AggregatedStats)

	for _, stats := range aggregatedStats {
		v, exists := values[stats.Type]
		if !exists {
			// Se non esiste ancora, inizializzo
			values[stats.Type] = types.AggregatedStats{
				Timestamp: timestamp.UTC().Unix(),
				Macrozone: environment.EdgeMacrozone,
				Min:       stats.Min,
				Max:       stats.Max,
				Sum:       stats.Sum,
				Count:     stats.Count,
			}
		} else {
			// Aggiorno i valori esistenti
			if stats.Min < v.Min {
				v.Min = stats.Min
			}
			if stats.Max > v.Max {
				v.Max = stats.Max
			}
			v.Sum += stats.Sum
			v.Count += stats.Count
			values[stats.Type] = v
		}
	}

	// Ora creo le statistiche aggregate per la macrozona
	macrozoneStats := make([]types.AggregatedStats, 0, len(values))
	for sensorType, v := range values {
		avg := 0.0
		if v.Count > 0 {
			avg = v.Sum / float64(v.Count)
		}
		macrozoneStats = append(macrozoneStats, types.AggregatedStats{
			Timestamp: v.Timestamp,
			Macrozone: v.Macrozone,
			Type:      sensorType,
			Min:       v.Min,
			Max:       v.Max,
			Avg:       avg,
			Sum:       v.Sum,
			Count:     v.Count,
		})
	}
	return macrozoneStats

}
