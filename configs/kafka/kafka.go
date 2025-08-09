package kafka

const BROKER = "localhost"
const PORT = "9094"

// Topics
// I topic permettono le comunicazioni tra i vari hub.
// In particolare:
// - il topic edge-hub permette la comunicazione tra l'edge hub e il proximity fog hub.
// - il topic proximity-fog-hub permette la comunicazione tra il proximity fog hub e l'intermediate fog hub.
// - il topic intermediate-fog-hub permette la comunicazione tra l'intermediate fog hub e il cloud.
// Il nome e la relativa costante deriva da chi sta scrivendo il messaggio.

// EDGE_HUB_TOPIC permette la comunicazione tra l'edge hub e il proximity fog hub.
const EDGE_HUB_TOPIC = "raw-data-edge-hub"

// PROXIMITY_FOG_HUB_DATA_TOPIC permette la comunicazione tra il proximity fog hub e l'intermediate fog hub per lo scambio dei dati.
const PROXIMITY_FOG_HUB_DATA_TOPIC = "aggregated-data-proximity-fog-hub"

// PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC permette la comunicazione tra il proximity fog hub e l'intermediate fog hub per lo scambio dei messaggi di comunicazione.
const PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC = "configuration-proximity-fog-hub"

// INTERMEDIATE_FOG_HUB_TOPIC permette la comunicazione tra l'intermediate fog hub e il cloud.
const INTERMEDIATE_FOG_HUB_TOPIC = "persistence-data-intermediate-fog-hub"

// AGGREGATED_STATS_TOPIC trasporta le statistiche aggregate (min, max, media)
// calcolate dal Proximity-Fog-Hub ogni tot minuti.
// Questo topic è separato per distinguere i dati riassuntivi da quelli in tempo reale.
const AGGREGATED_STATS_TOPIC = "statistics-data-proximity-fog-hub"
