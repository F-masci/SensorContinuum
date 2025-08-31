CREATE EXTENSION IF NOT EXISTS timescaledb;

-- =========================================================
-- ===========  TABELLA PER CACHE SENSOR DATA ==============
-- =========================================================

-- 1. Creiamo la tabella per la cache locale del proximity-hub
CREATE TABLE IF NOT EXISTS sensor_measurements_cache (
    time            TIMESTAMPTZ       NOT NULL,
    macrozone_name  VARCHAR(255)      NOT NULL,
    zone_name       VARCHAR(255)      NOT NULL,
    sensor_id       VARCHAR(255)      NOT NULL,
    type            VARCHAR(50)       NOT NULL,
    value           DOUBLE PRECISION  NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent')),
    PRIMARY KEY (time, macrozone_name, zone_name, sensor_id, type)
);

-- 2. La trasformiamo in un'hypertable, partizionata per tempo sulla colonna 'time'
SELECT create_hypertable('sensor_measurements_cache', 'time', if_not_exists => TRUE, chunk_time_interval => interval '1 hour');

-- 3. Impostiamo una politica di retention per cancellare dati più vecchi di 1 giorno
SELECT add_retention_policy('sensor_measurements_cache', INTERVAL '1 days');

-- ===========================================================
-- ===========  TABELLA PER CACHE AGGREGATED STATS ===========
-- ===========================================================

-- 1. Creiamo la tabella per la cache locale del proximity-hub
-- NOTA: zone_name di default è stringa vuota per le statistiche a livello di macrozona
-- e non può essere NULL per la chiave primaria (che evita duplicati)
CREATE TABLE IF NOT EXISTS aggregated_stats_cache (
    time            TIMESTAMPTZ       NOT NULL,
    zone_name       TEXT              NOT NULL DEFAULT '',
    type            TEXT              NOT NULL,
    min_value       DOUBLE PRECISION  NOT NULL,
    max_value       DOUBLE PRECISION  NOT NULL,
    avg_value       DOUBLE PRECISION  NOT NULL,
    avg_sum         DOUBLE PRECISION  NOT NULL,
    avg_count       INTEGER           NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent')),
    PRIMARY KEY (time, zone_name, type)
);

-- 2. La trasformiamo in un'hypertable, partizionata per tempo sulla colonna 'time'
SELECT create_hypertable('aggregated_stats_cache', 'time', if_not_exists => TRUE, chunk_time_interval => interval '4 hour');

-- 3. Impostiamo una politica di retention per cancellare dati più vecchi di 2 giorno
SELECT add_retention_policy('aggregated_stats_cache', INTERVAL '2 days');