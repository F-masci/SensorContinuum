CREATE EXTENSION IF NOT EXISTS timescaledb;

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
SELECT create_hypertable('sensor_measurements', 'time', if_not_exists => TRUE);

-- 3. Crea indice per ottimizzare le query
CREATE INDEX IF NOT EXISTS idx_sensor_id ON sensor_measurements (macrozone_name, zone_name, sensor_id);

-- 4. Aggiungi indice composito utile per filtro e aggregazione per tipo
CREATE INDEX IF NOT EXISTS idx_sensor_type_time ON sensor_measurements (type, time DESC);

-- 5. (Opzionale) Partizionamento secondario per scaling orizzontale futuro
-- SELECT create_hypertable('sensor_measurements', 'time', chunk_time_interval => interval '1 day');

-- ==========================================================================
-- ======== TABELLA PER STATISTICHE AGGREGATE A LIVELLO DI MACROZONA ========
-- ==========================================================================

CREATE TABLE IF NOT EXISTS region_aggregated_statistics (
    time            TIMESTAMPTZ       NOT NULL,
    type            TEXT              NOT NULL,
    min_value       DOUBLE PRECISION  NOT NULL,
    max_value       DOUBLE PRECISION  NOT NULL,
    avg_value       DOUBLE PRECISION  NOT NULL,
    avg_sum         DOUBLE PRECISION  NOT NULL,
    avg_count       INTEGER           NOT NULL
);

SELECT create_hypertable('region_aggregated_statistics', 'time', if_not_exists => TRUE);

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

SELECT create_hypertable('macrozone_aggregated_statistics', 'time', if_not_exists => TRUE);

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

SELECT create_hypertable('zone_aggregated_statistics', 'time', if_not_exists => TRUE);