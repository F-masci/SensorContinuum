package kafka

const BROKER = "localhost"
const PORT = "9094"

// Topics
// I topic permettono le comunicazioni tra i vari hub.
// In particolare:
// - il topic edge-hub permette la comunicazione tra l'edge hub e il proximity fog hub.
// - il topic proximity-fog-hub permette la comunicazione tra il proximity fog hub e l'intermediate fog hub.
// Il nome e la relativa costante deriva da chi sta scrivendo il messaggio.
const EDGE_HUB_TOPIC = "raw-data-edge-hub"
const PROXIMITY_FOG_HUB_TOPIC = "aggregated-data-proximity-fog-hub"
