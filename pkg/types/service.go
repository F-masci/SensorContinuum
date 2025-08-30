package types

// Service Definizione dei servizi
type Service string

const (
	// SensorAgentService servizio che gira sui sensori per simulare l'invio dei dati
	SensorAgentService Service = "sensor_agent"

	// EdgeHubService servizio completo dell' edge-hub
	EdgeHubService Service = "edge_hub"
	// EdgeHubConfigurationService si occupa di gestire i messaggi di configurazione
	EdgeHubConfigurationService Service = "edge_hub_configuration"
	// EdgeHubFilterService si occupa di filtrare e salvare in cache i dati in arrivo dai sensori
	EdgeHubFilterService Service = "edge_hub_filter"
	// EdgeHubAggregatorService si occupa di aggregare i dati e inviarli al proximity-fog-hub
	EdgeHubAggregatorService Service = "edge_hub_aggregator"
	// EdgeHubCleanerService si occupa di pulire la cache locale e notificare attivamente i sensori offline
	EdgeHubCleanerService Service = "edge_hub_cleaner"

	// ProximityHubService servizio completo del proximity-fog-hub
	ProximityHubService Service = "proximity_hub"
	// ProximityHubLocalCacheService si occupa di gestire il salvataggio in cache locale
	ProximityHubLocalCacheService Service = "proximity_hub_local_cache"
	// ProximityHubConfigurationService si occupa di gestire i messaggi di configurazione
	ProximityHubConfigurationService Service = "proximity_hub_configuration"
	// ProximityHubHeartbeatService si occupa di gestire i messaggi di heartbeat
	ProximityHubHeartbeatService Service = "proximity_hub_heartbeat"
	// ProximityHubAggregatorService si occupa di aggregare i dati
	ProximityHubAggregatorService Service = "proximity_hub_aggregator"
	// ProximityHubDispatcherService si occupa di inviare i dati aggregati al Intermediate Fog Hub
	ProximityHubDispatcherService Service = "proximity_hub_dispatcher"
	// ProximityHubCleanerService si occupa di pulire la cache locale
	ProximityHubCleanerService Service = "proximity_hub_cleaner"

	// IntermediateHubService servizio completo dell' intermediate-fog-hub
	IntermediateHubService Service = "intermediate_hub"
	// IntermediateHubRealtimeService si occupa di gestire i dati in tempo reale
	IntermediateHubRealtimeService Service = "intermediate_hub_realtime"
	// IntermediateHubStatisticsService si occupa di gestire i dati statistici
	IntermediateHubStatisticsService Service = "intermediate_hub_statistics"
	// IntermediateHubConfigurationService si occupa di gestire i messaggi di configurazione
	IntermediateHubConfigurationService Service = "intermediate_hub_configuration"
	// IntermediateHubHeartbeatService si occupa di gestire i messaggi di heartbeat
	IntermediateHubHeartbeatService Service = "intermediate_hub_heartbeat"
	// IntermediateHubAggregatorService si occupa di calcolare e salvare i dati aggregati nel database centrale
	IntermediateHubAggregatorService Service = "intermediate_hub_aggregator"
)

// OperationModeType Definizione dei tipi di modalità operativa
type OperationModeType string

const (
	// OperationModeLoop per eseguire in un ciclo continuo il servizio (default)
	OperationModeLoop OperationModeType = "loop"
	// OperationModeOnce per eseguire una singola iterazione del servizio
	OperationModeOnce OperationModeType = "once"
)
