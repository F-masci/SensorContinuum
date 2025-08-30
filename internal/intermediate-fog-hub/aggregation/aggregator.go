package aggregation

import (
	"SensorContinuum/internal/intermediate-fog-hub/environment"
	"SensorContinuum/internal/intermediate-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"os"
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
// per la regione, nell'intervallo di tempo specificato.
// Le statistiche vengono poi salvate nella tabella relativa.
func AggregateSensorData(ctx context.Context) {

	// Stabilisce la connessione al database dei sensori.
	err := storage.SetupSensorDbConnection()
	if err != nil {
		logger.Log.Error("Failed to connect to the sensor database: ", err)
		os.Exit(1)
	}

	// 1. Calcola gli intervalli allineati per l'aggregazione
	lastAggregation, err := storage.GetLastRegionAggregatedData(ctx)
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

		// 2. Esegui la query per ottenere le statistiche dei dati arrivati nell'intervallo alignedStartTime e alignedEndTime
		stats, err := storage.GetMacrozoneStatisticsData(ctx, alignedStartTime, alignedEndTime)
		if err != nil {
			logger.Log.Error("Failed to calculate periodic statistics, error: ", err)
			return
		}

		if len(stats) == 0 {
			logger.Log.Info("No data to send, skipping aggregation")
			return
		}

		// 3. Salva le statistiche aggregate
		values := make(map[string]types.AggregatedStats)
		for _, stat := range stats {
			v, exists := values[stat.Type]
			if !exists {
				// Se non esiste ancora, inizializzo
				values[stat.Type] = types.AggregatedStats{
					Timestamp: alignedEndTime.UTC().Unix(),
					Type:      stat.Type,
					Min:       stat.Min,
					Max:       stat.Max,
					Sum:       stat.Sum,
					Count:     stat.Count,
				}
			} else {
				// Aggiorno i valori esistenti
				if stat.Min < v.Min {
					v.Min = stat.Min
				}
				if stat.Max > v.Max {
					v.Max = stat.Max
				}
				v.Sum += stat.Sum
				v.Count += stat.Count
				values[stat.Type] = v
			}
		}

		// Salva le statistiche aggregate a livello di region nel database
		for _, agg := range values {
			if agg.Count > 0 {
				agg.Avg = agg.Sum / float64(agg.Count)
			} else {
				agg.Avg = 0
			}
			logger.Log.Debug("Aggregated region data: ", agg)
			if err := storage.InsertRegionStatisticsData(agg); err != nil {
				logger.Log.Error("Failed to save region aggregated data, type: ", agg.Type, ", error: ", err)
				continue
			}
			logger.Log.Info("Macrozone region data saved successfully, type: ", agg.Type)
		}
	}

}
