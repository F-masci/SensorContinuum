CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 1. Crea la tabella base
CREATE TABLE IF NOT EXISTS sensor_measurements (
    time         TIMESTAMPTZ       NOT NULL,
    building_id  TEXT              NOT NULL,
    floor_id     TEXT              NOT NULL,
    sensor_id    TEXT              NOT NULL,
    type         TEXT              NOT NULL,
    value        DOUBLE PRECISION  NOT NULL
);

-- 2. Crea la hypertable (solo se non esiste già)
SELECT create_hypertable('sensor_measurements', 'time', if_not_exists => TRUE);

-- 3. Crea indice per ottimizzare le query su sensor_id (singolo)
CREATE INDEX IF NOT EXISTS idx_sensor_id ON sensor_measurements (sensor_id);

-- 4. Aggiungi indice composito utile per filtro e aggregazione per tipo
CREATE INDEX IF NOT EXISTS idx_sensor_type_time ON sensor_measurements (type, time DESC);

-- 5. (Opzionale) Partizionamento secondario per scaling orizzontale futuro
-- SELECT create_hypertable('sensor_measurements', 'time', chunk_time_interval => interval '1 day');

