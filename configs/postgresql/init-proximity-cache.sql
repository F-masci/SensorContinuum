CREATE EXTENSION IF NOT EXISTS timescaledb;

-- ===============================================================
-- ===========  TABELLA PER CACHE PROXIMITY-FOG-HUB ==============
-- ===============================================================

-- 1. Creiamo la tabella per la cache locale del proximity-hub
CREATE TABLE sensor_measurements_cache (
    time            TIMESTAMPTZ       NOT NULL,
    macrozone_name  VARCHAR(255)      NOT NULL,
    zone_name       VARCHAR(255)      NOT NULL,
    sensor_id       VARCHAR(255)      NOT NULL,
    type            VARCHAR(50)       NOT NULL,
    value           DOUBLE PRECISION  NOT NULL
);

-- 2. La trasformiamo in un'hypertable, partizionata per tempo sulla colonna 'time'
SELECT create_hypertable('sensor_measurements_cache', 'time');

-- 3. Impostiamo una politica di retention per cancellare dati più vecchi di 1 giorno
SELECT add_retention_policy('sensor_measurements_cache', INTERVAL '1 days');

CREATE TABLE IF NOT EXISTS aggregated_stats_outbox (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Chiave primaria univoca per ogni messaggio
    payload         JSONB NOT NULL,                             -- Il contenuto del messaggio (le statistiche in formato JSON)
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',     -- Stato del messaggio: 'pending' o 'sent'
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),         -- Timestamp di creazione del record
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()          -- Timestamp dell'ultimo aggiornamento
);

-- Indice sullo stato per velocizzare la ricerca dei messaggi da inviare.
CREATE INDEX IF NOT EXISTS idx_aggregated_stats_outbox_status ON aggregated_stats_outbox(status);