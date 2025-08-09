package types

type Service string

const (
	SensorAgentService          Service = "sensor_agent"
	EdgeHubService              Service = "edge_hub"
	EdgeHubConfigurationService Service = "edge_hub_configuration"
	EdgeHubFilterService        Service = "edge_hub_filter"
	EdgeHubAggregatorService    Service = "edge_hub_aggregator"
	EdgeHubCleanerService       Service = "edge_hub_cleaner"
	ProximityHubService         Service = "proximity_hub"
	IntrermediateHubService     Service = "intermediate_hub"
)
