package environment

import "time"

const (
	// AggregatedDataCutOff definisce il tempo massimo di validità dei dati aggregati.
	// Se i dati sono più vecchi di questo valore, non vengono considerati validi
	AggregatedDataCutOff time.Duration = 2 * time.Hour

	// YearlyVariationOffset definisce l'offset temporale per il calcolo della variazione annuale.
	// Viene sottratto alla data corrente per ottenere la data di cui calcolare la variazione
	YearlyVariationOffset time.Duration = -24 * time.Hour
)
