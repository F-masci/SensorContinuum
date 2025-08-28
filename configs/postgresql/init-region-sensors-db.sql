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
SELECT create_hypertable('sensor_measurements', 'time', if_not_exists => TRUE, chunk_time_interval => interval '1 day');

-- 3. Crea indice per ottimizzare le query
CREATE INDEX IF NOT EXISTS idx_sensor_id ON sensor_measurements (macrozone_name, zone_name, sensor_id);

-- 4. Aggiungi indice composito utile per filtro e aggregazione per tipo
CREATE INDEX IF NOT EXISTS idx_sensor_type_time ON sensor_measurements (type, time DESC);

-- 5. (Opzionale) Partizionamento secondario per scaling orizzontale futuro
-- SELECT create_hypertable('sensor_measurements', 'time', chunk_time_interval => interval '1 day');

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

-- ============================================================
-- ======== VISTE CONTINUOUS AGGREGATES PER MACROZONE ========
-- ============================================================

-- Aggregazione giornaliera per macrozona
CREATE MATERIALIZED VIEW IF NOT EXISTS macrozone_daily_agg
            WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS day,
    macrozone_name,
    type,
    MIN(min_value) AS min_value,
    MAX(max_value) AS max_value,
    SUM(avg_sum) / NULLIF(SUM(avg_count),0) AS avg_value,
    SUM(avg_sum) AS total_sum,
    SUM(avg_count) AS total_count
FROM macrozone_aggregated_statistics
GROUP BY day, macrozone_name, type;

-- Aggregazione settimanale per macrozona
CREATE MATERIALIZED VIEW IF NOT EXISTS macrozone_weekly_agg
            WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 week', time) AS week,
    macrozone_name,
    type,
    MIN(min_value) AS min_value,
    MAX(max_value) AS max_value,
    SUM(avg_sum) / NULLIF(SUM(avg_count),0) AS avg_value,
    SUM(avg_sum) AS total_sum,
    SUM(avg_count) AS total_count
FROM macrozone_aggregated_statistics
GROUP BY week, macrozone_name, type;

-- Aggregazione mensile per macrozona
CREATE MATERIALIZED VIEW IF NOT EXISTS macrozone_monthly_agg
            WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 month', time) AS month,
    macrozone_name,
    type,
    MIN(min_value) AS min_value,
    MAX(max_value) AS max_value,
    SUM(avg_sum) / NULLIF(SUM(avg_count),0) AS avg_value,
    SUM(avg_sum) AS total_sum,
    SUM(avg_count) AS total_count
FROM macrozone_aggregated_statistics
GROUP BY month, macrozone_name, type;

-- Aggregazione annuale per macrozona
CREATE MATERIALIZED VIEW IF NOT EXISTS macrozone_yearly_agg
            WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 year', time) AS year,
    macrozone_name,
    type,
    MIN(min_value) AS min_value,
    MAX(max_value) AS max_value,
    SUM(avg_sum) / NULLIF(SUM(avg_count),0) AS avg_value,
    SUM(avg_sum) AS total_sum,
    SUM(avg_count) AS total_count
FROM macrozone_aggregated_statistics
GROUP BY year, macrozone_name, type;

-- ============================================================
-- ======== VISTE CONTINUOUS AGGREGATES PER REGIONE ===========
-- ============================================================

-- Aggregazione giornaliera per regione
CREATE MATERIALIZED VIEW IF NOT EXISTS region_daily_agg
            WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS day,
    type,
    MIN(min_value) AS min_value,
    MAX(max_value) AS max_value,
    SUM(avg_sum) / NULLIF(SUM(avg_count),0) AS avg_value,
    SUM(avg_sum) AS total_sum,
    SUM(avg_count) AS total_count
FROM region_aggregated_statistics
GROUP BY day, type;

-- Aggregazione settimanale per regione
CREATE MATERIALIZED VIEW IF NOT EXISTS region_weekly_agg
            WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 week', time) AS week,
    type,
    MIN(min_value) AS min_value,
    MAX(max_value) AS max_value,
    SUM(avg_sum) / NULLIF(SUM(avg_count),0) AS avg_value,
    SUM(avg_sum) AS total_sum,
    SUM(avg_count) AS total_count
FROM region_aggregated_statistics
GROUP BY week, type;

-- Aggregazione mensile per regione
CREATE MATERIALIZED VIEW IF NOT EXISTS region_monthly_agg
            WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 month', time) AS month,
    type,
    MIN(min_value) AS min_value,
    MAX(max_value) AS max_value,
    SUM(avg_sum) / NULLIF(SUM(avg_count),0) AS avg_value,
    SUM(avg_sum) AS total_sum,
    SUM(avg_count) AS total_count
FROM region_aggregated_statistics
GROUP BY month, type;

-- Aggregazione annuale per regione
CREATE MATERIALIZED VIEW IF NOT EXISTS region_yearly_agg
            WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 year', time) AS year,
    type,
    MIN(min_value) AS min_value,
    MAX(max_value) AS max_value,
    SUM(avg_sum) / NULLIF(SUM(avg_count),0) AS avg_value,
    SUM(avg_sum) AS total_sum,
    SUM(avg_count) AS total_count
FROM region_aggregated_statistics
GROUP BY year, type;

-- Macrozone
SELECT add_continuous_aggregate_policy('macrozone_daily_agg', start_offset=>'2 days', end_offset=>'0 hours', schedule_interval=>'1 hour');
SELECT add_continuous_aggregate_policy('macrozone_weekly_agg', start_offset=>'2 weeks', end_offset=>'0 hours', schedule_interval=>'6 hours');
SELECT add_continuous_aggregate_policy('macrozone_monthly_agg', start_offset=>'2 months', end_offset=>'0 hours', schedule_interval=>'12 hours');
SELECT add_continuous_aggregate_policy('macrozone_yearly_agg', start_offset=>'2 years', end_offset=>'0 hours', schedule_interval=>'1 day');

-- Regione
SELECT add_continuous_aggregate_policy('region_daily_agg', start_offset=>'2 days', end_offset=>'0 hours', schedule_interval=>'1 hour');
SELECT add_continuous_aggregate_policy('region_weekly_agg', start_offset=>'2 weeks', end_offset=>'0 hours', schedule_interval=>'6 hours');
SELECT add_continuous_aggregate_policy('region_monthly_agg', start_offset=>'2 months', end_offset=>'0 hours', schedule_interval=>'12 hours');
SELECT add_continuous_aggregate_policy('region_yearly_agg', start_offset=>'2 years', end_offset=>'0 hours', schedule_interval=>'1 day');

