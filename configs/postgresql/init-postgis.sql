-- Abilita estensioni necessarie
CREATE EXTENSION IF NOT EXISTS postgis;

-- 1. Tabella regioni
CREATE TABLE IF NOT EXISTS regions (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- 2. Tabella edifici con coordinate geografiche (Point)
CREATE TABLE IF NOT EXISTS buildings (
    id         SERIAL PRIMARY KEY,
    region_id  INTEGER NOT NULL REFERENCES regions(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    location   geometry(Point, 4326) NOT NULL
);

-- Indice spaziale per ricerche geografiche
CREATE INDEX IF NOT EXISTS idx_buildings_location ON buildings USING GIST (location);