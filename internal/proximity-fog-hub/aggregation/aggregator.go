package aggregation

import (
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"time"
)

// PerformAggregationAndSend è la funzione che viene eseguita periodicamente ogni tot minuti stabiliti
// l'idea non è quella di ricevere i dati degli ultimi tot minuti in maniera casuale ma di restituire i dati
// compresi in un intervallo temporale che parte da uno start e termina in un end (ovviamente questo
// intervallo durerà quei tot minuti che ci siamo stabiliti)
// salva nella tabella 'outbox' le statistiche e un processo 'dispatcher' separato si occuperà poi dell invio

func PerformAggregationAndSend() {
	logger.Log.Info("Execution of periodic aggregation started")
	ctx := context.Background()

	// calcola il periodo di 2 minuti rispetto
	// a 5 minuti fa, in modo da avere anche i
	// dati che hanno subito un ritardo
	now := time.Now().UTC().Add(-5 * time.Minute)
	//allineo il tempo attuale al limite di minuti
	//quindi se ora sono le 15:48:30, alignedEndTime sarà 15:44:00
	intervalDuration := 5 * time.Minute
	alignedEndTime := now.Truncate(intervalDuration)
	//l'inizio del periodo è di 5 minuti prima di alignedEndTime
	alignedStartTime := alignedEndTime.Add(-intervalDuration)
	logger.Log.Info("Aggregation data for interval, start_time: ", alignedStartTime.Format(time.RFC3339), ", end_time: ", alignedEndTime.Format(time.RFC3339))

	// 2. Esegui la query per ottenere le statistiche dei dati arrivati nell'intervallo alignedStartTime e alignedEndTime
	stats, err := storage.GetZoneAggregatedData(ctx, alignedStartTime, alignedEndTime)
	if err != nil {
		logger.Log.Error("Failed to calculate periodic statistics, error: ", err)
		return
	}

	if len(stats) == 0 {
		logger.Log.Info("No data to send, skipping aggregation")
		return
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

		if err := storage.InsertAggregatedStatsOutbox(ctx, stat); err != nil {
			logger.Log.Error("Failure to save statistics to outbox, type: ", stat.Type, ", error: ", err)
			// Non ci fermiamo, proviamo a salvare le altre
			continue
		}
		logger.Log.Info("Statistics successfully saved to outbox for type: ", stat.Type)
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
