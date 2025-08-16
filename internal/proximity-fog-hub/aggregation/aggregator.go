package aggregation

import (
	"SensorContinuum/internal/proximity-fog-hub/comunication"
	"SensorContinuum/internal/proximity-fog-hub/environment"
	"SensorContinuum/internal/proximity-fog-hub/storage"
	"SensorContinuum/pkg/logger"
	"context"
	"time"
)

// PerformAggregationAndSend è la funzione che viene eseguita periodicamente ogni tot minuti stabiliti
// l'idea non è quella di ricevere i dati degli ultimi tot minuti in maniera casuale ma di restituire i dati
// compresi in un intervallo temporale che parte da uno start e termina in un end ( e ovviamente questo
// intervallo durerà quei tot minuti che ci siamo stabiliti ( in questo caso 2)

func PerformAggregationAndSend() {
	logger.Log.Info("execution of periodic aggregation started")
	ctx := context.Background()

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
	logger.Log.Info("aggregation data for interval, start_time: ", alignedStartTime.Format(time.RFC3339), ", end_time: ", alignedEndTime.Format(time.RFC3339))

	// 2. Esegui la query per ottenere le statistiche dei dati arrivati nell'intervallo alignedStartTime e alignedEndTime
	stats, err := storage.GetValueToSend(ctx, alignedStartTime, alignedEndTime)
	if err != nil {
		logger.Log.Error("failed to calculate periodic statistics, error: ", err)
		return
	}

	if len(stats) == 0 {
		logger.Log.Info("No data to send, skipping aggregation")
		return
	}

	// 3. Invia ogni statistica a Kafka
	for _, stat := range stats {
		// Arricchiamo la statistica con dati contestuali
		stat.Timestamp = alignedEndTime.UTC().Unix()
		stat.Macrozone = environment.EdgeMacrozone

		logger.Log.Info("Statistics calculated for the type: ", stat.Type, ", min: ", stat.Min, ", max: ", stat.Max, ", avg: ", stat.Avg)

		if err := comunication.SendAggregatedData(stat); err != nil {
			logger.Log.Error("Failure to send statistics to Kafka, type: ", stat.Type, ", error: ", err)
			// Non ci fermiamo, proviamo a inviare le altre
			continue
		}
		logger.Log.Info("Statistics successfully sent to Kafka for type: ", stat.Type)
	}
}
