package main

import (
	regionAPI "SensorContinuum/internal/api-backend/region"
	"context"
	"time"
)

func main() {

	ctx := context.Background()

	starts := make([]time.Time, 0)
	ends := make([]time.Time, 0)

	starts = append(starts, time.Now().Add(-1*time.Hour))
	ends = append(ends, time.Now().Add(10*time.Minute))

	for i := 0; i < len(starts) && i < len(ends); i++ {
		start := starts[i]
		end := ends[i]

		regionAggregated, err := regionAPI.ComputeAggregatedSensorData(ctx, "region-001", start, end)
		if err != nil {
			panic(err)
		}
		for _, agg := range regionAggregated {
			println("Type:", agg.Type, "Min:", agg.Min, "Max:", agg.Max, "Avg:", agg.Avg, "Sum:", agg.Sum, "Count:", agg.Count)
			// err = regionAPI.SaveAggregatedSensorData(ctx, "region-001", agg, end)
			if err != nil {
				panic(err)
			}
		}
	}

}
