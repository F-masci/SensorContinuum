CREATE EXTENSION IF NOT EXISTS timescaledb;


CREATE EXTENSION IF NOT EXISTS timescaledb;

-- =================================================================================
-- ===========   TABELLA CACHE PER DATI REAL-TIME (PROXIMITY-HUB) ====================
-- =================================================================================
-- Questa tabella agisce come cache locale per il proximity-hub, conservando i dati
-- recenti su cui vengono calcolate le statistiche periodiche.
-- NOTA: Questa tabella è utilizzata dal proximity-hub ma viene creata in questo
-- schema condiviso su richiesta esplicita.

CREATE TABLE IF NOT EXISTS sensor_measurements_cache (
                                                         time            TIMESTAMPTZ       NOT NULL,
                                                         macrozone_name  VARCHAR(255)      NOT NULL,
    zone_name       VARCHAR(255)      NOT NULL,
    sensor_id       VARCHAR(255)      NOT NULL,
    type            VARCHAR(50)       NOT NULL,
    value           DOUBLE PRECISION  NOT NULL
    );

-- La trasformiamo in un'hypertable, partizionata per tempo
SELECT create_hypertable('sensor_measurements_cache', 'time', if_not_exists => TRUE);

-- Impostiamo una politica di retention per cancellare dati più vecchi di 1 giorno
SELECT add_retention_policy('sensor_measurements_cache', INTERVAL '1 days');

-- ============================================================
-- ======== TABELLA PRINCIAPALE PER I DATI DEI SENSORI ========
-- ============================================================

-- 1. Crea la tabella base
CREATE TABLE IF NOT EXISTS sensor_measurements (
    time            TIMESTAMPTZ       NOT NULL,
    macrozone_name  TEXT              NOT NULL,
    zone_name       TEXT              NOT NULL,
    sensor_id       TEXT              NOT NULL,
    type            TEXT              NOT NULL,
    value           DOUBLE PRECISION  NOT NULL
);

-- 2. Crea la hypertable (solo se non esiste già)
SELECT create_hypertable('sensor_measurements', 'time', if_not_exists => TRUE, chunk_time_interval => interval '1 day');

-- 3. Crea indice per ottimizzare le query
CREATE INDEX IF NOT EXISTS idx_sensor_id ON sensor_measurements (macrozone_name, zone_name, sensor_id);

-- 4. Aggiungi indice composito utile per filtro e aggregazione per tipo
CREATE INDEX IF NOT EXISTS idx_sensor_type_time ON sensor_measurements (type, time DESC);

-- ========================================================================
-- ======== TABELLA PER STATISTICHE AGGREGATE A LIVELLO DI REGIONE ========
-- ========================================================================

-- 1. Statistiche aggregate a livello di regione
CREATE TABLE IF NOT EXISTS region_aggregated_statistics (
    time            TIMESTAMPTZ       NOT NULL,
    type            TEXT              NOT NULL,
    min_value       DOUBLE PRECISION  NOT NULL,
    max_value       DOUBLE PRECISION  NOT NULL,
    avg_value       DOUBLE PRECISION  NOT NULL,
    avg_sum         DOUBLE PRECISION  NOT NULL,
    avg_count       INTEGER           NOT NULL
);

-- 2. Crea la hypertable (solo se non esiste già)
SELECT create_hypertable('region_aggregated_statistics', 'time', if_not_exists => TRUE);

-- ==========================================================================
-- ======== TABELLA PER STATISTICHE AGGREGATE A LIVELLO DI MACROZONA ========
-- ==========================================================================

-- 1. Statistiche aggregate a livello di macrozona
CREATE TABLE IF NOT EXISTS macrozone_aggregated_statistics (
    time            TIMESTAMPTZ       NOT NULL,
    macrozone_name  TEXT              NOT NULL,
    type            TEXT              NOT NULL,
    min_value       DOUBLE PRECISION  NOT NULL,
    max_value       DOUBLE PRECISION  NOT NULL,
    avg_value       DOUBLE PRECISION  NOT NULL,
    avg_sum         DOUBLE PRECISION  NOT NULL,
    avg_count       INTEGER           NOT NULL
);

-- 2. Crea la hypertable (solo se non esiste già)
SELECT create_hypertable('macrozone_aggregated_statistics', 'time', if_not_exists => TRUE);

-- =====================================================================
-- ======== TABELLA PER STATISTICHE AGGREGATE A LIVELLO DI ZONA ========
-- =====================================================================

-- 1. Statistiche aggregate a livello di zona
CREATE TABLE IF NOT EXISTS zone_aggregated_statistics (
    time            TIMESTAMPTZ       NOT NULL,
    macrozone_name  TEXT              NOT NULL,
    zone_name       TEXT              NOT NULL,
    type            TEXT              NOT NULL,
    min_value       DOUBLE PRECISION  NOT NULL,
    max_value       DOUBLE PRECISION  NOT NULL,
    avg_value       DOUBLE PRECISION  NOT NULL,
    avg_sum         DOUBLE PRECISION  NOT NULL,
    avg_count       INTEGER           NOT NULL
);

-- 2. Crea la hypertable (solo se non esiste già)
SELECT create_hypertable('zone_aggregated_statistics', 'time', if_not_exists => TRUE);

CREATE TABLE IF NOT EXISTS aggregated_stats_outbox (
                                                       id              UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Chiave primaria univoca per ogni messaggio
    payload         JSONB NOT NULL,                             -- Il contenuto del messaggio (le statistiche in formato JSON)
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',     -- Stato del messaggio: 'pending' o 'sent'
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),         -- Timestamp di creazione del record
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()          -- Timestamp dell'ultimo aggiornamento
    );

-- Indice sullo stato per velocizzare la ricerca dei messaggi da inviare.
CREATE INDEX IF NOT EXISTS idx_aggregated_stats_outbox_status ON aggregated_stats_outbox(status);