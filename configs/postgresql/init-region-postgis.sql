-- Abilita estensioni necessarie
CREATE EXTENSION IF NOT EXISTS postgis;

-- 1. Tabella edifici con coordinate geografiche
CREATE TABLE IF NOT EXISTS buildings (
    id                  SERIAL PRIMARY KEY,
    name                TEXT NOT NULL,
    location            geometry(Point, 4326) NOT NULL,
    registration_time   TIMESTAMP,
    last_comunication   TIMESTAMP
);

-- Indice spaziale per ricerche geografiche
CREATE INDEX IF NOT EXISTS idx_buildings_location ON buildings USING GIST (location);

-- 2. Tabella piani
CREATE TABLE IF NOT EXISTS floors (
    id                SERIAL PRIMARY KEY,
    building_id       INTEGER NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    lastComunication  TIMESTAMP
);

-- 3. Tabella sensori
CREATE TABLE IF NOT EXISTS sensors (
    id                SERIAL PRIMARY KEY,
    floor_id          INTEGER NOT NULL REFERENCES floors(id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    lastComunication  TIMESTAMP
);