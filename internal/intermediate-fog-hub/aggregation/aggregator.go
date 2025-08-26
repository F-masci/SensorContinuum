package aggregation

import (
	"SensorContinuum/internal/intermediate-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"os"
	"time"
)

func PerformAggregationAndSave() {
	logger.Log.Info("Execution of periodic aggregation started")
	ctx := context.Background()

	err := storage.SetupSensorDbConnection()
	if err != nil {
		logger.Log.Error("Failed to connect to the sensor database: ", err)
		os.Exit(1)
	}

	// calcola il periodo di 2 minuti rispetto
	// a 5 minuti fa, in modo da avere anche i
	// dati che hanno subito un ritardo
	now := time.Now().UTC().Add(-5 * time.Minute)
	//allineo il tempo attuale al limite di 2 minuti
	//quindi se ora sono le 15:48:30, alignedEndTime sarà 15:47:00
	intervalDuration := 2 * time.Minute
	alignedEndTime := now.Truncate(intervalDuration)
	//l'inizio del periodo è di 2 minuti prima di alignedEndTime
	alignedStartTime := alignedEndTime.Add(-intervalDuration)
	logger.Log.Info("Aggregation data for interval, start_time: ", alignedStartTime.Format(time.RFC3339), ", end_time: ", alignedEndTime.Format(time.RFC3339))

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
