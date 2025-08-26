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

-- 2. La trasformiamo in un'hypertable, partizionata per tempo
SELECT create_hypertable('sensor_measurements_cache', 'time');

-- 3. Impostiamo una politica di retention per cancellare dati più vecchi di 1 giorno
SELECT add_retention_policy('sensor_measurements_cache', INTERVAL '1 days');