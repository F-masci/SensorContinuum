package timeouts

import "time"

// IsAliveSensorTimeout è il timeout per considerare un sensore "vivo"
const IsAliveSensorTimeout = 5 * time.Minute

// IsAliveHubTimeout è il timeout per considerare un Hub "vivo"
const IsAliveHubTimeout = IsAliveSensorTimeout

// HeartbeatInterval è l'intervallo di invio del messaggio di heartbeat
const HeartbeatInterval = 3 * time.Minute
