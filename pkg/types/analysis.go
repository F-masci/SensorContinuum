package types

// VariationResult rappresenta il risultato dell'analisi Year-over-Year (YoY) per una specifica macrozona e tipo di dato.
type VariationResult struct {
	Macrozone string  `json:"macrozone"`
	Type      string  `json:"type"`
	Current   float64 `json:"current"`
	Previous  float64 `json:"previous"`
	DeltaPerc float64 `json:"delta_perc"`
	Timestamp int64   `json:"timestamp"`
}

// MacrozoneAnomaly rappresenta un'anomalia rilevata in una macrozona rispetto ai suoi vicini.
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

// TrendSimilarityResult contiene i risultati del confronto tra macrozona e regione
type TrendSimilarityResult struct {
	MacrozoneName string  `json:"macrozone"` // nome della macrozona
	RegionName    string  `json:"region"`
	Type          string  `json:"type"` // tipo di dato (es. PM10, NO2, etc.)
	Correlation   float64 `json:"correlation"`
	SlopeMacro    float64 `json:"slope_macro"`
	SlopeRegion   float64 `json:"slope_region"`
	Divergence    float64 `json:"divergence"`

	// Serie utilizzate nei calcoli
	MacrozoneSeries []AggregatedStats `json:"macrozone_series"`
	RegionSeries    []AggregatedStats `json:"region_series"`
}
