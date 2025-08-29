package types

type Service string

// Definizione dei servizi
const (
	SensorAgentService          Service = "sensor_agent"
	EdgeHubService              Service = "edge_hub"
	EdgeHubConfigurationService Service = "edge_hub_configuration"
	EdgeHubFilterService        Service = "edge_hub_filter"
	EdgeHubAggregatorService    Service = "edge_hub_aggregator"
	EdgeHubCleanerService       Service = "edge_hub_cleaner"
	ProximityHubService         Service = "proximity_hub"
	IntermediateHubService      Service = "intermediate_hub"
)
