CREATE EXTENSION IF NOT EXISTS postgis;

-- 1. Tabella regioni con name come chiave primaria
CREATE TABLE IF NOT EXISTS regions (
    name                TEXT PRIMARY KEY,
    location            geometry(Polygon, 4326) NOT NULL,
    creation_time       TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_regions_location ON regions USING GIST (location);

-- 2. Tabella macrozone con chiave primaria composta
CREATE TABLE IF NOT EXISTS macrozones (
    region_name         TEXT NOT NULL REFERENCES regions(name) ON DELETE CASCADE,
    name                TEXT NOT NULL,
    location            geometry(Polygon, 4326) NOT NULL,
    type                TEXT NOT NULL CHECK (type IN ('indoor', 'outdoor')),
    creation_time       TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (region_name, name)
);

CREATE INDEX IF NOT EXISTS idx_macrozones_location ON macrozones USING GIST (location);

-- 3. Tabella zone con chiave primaria composta
CREATE TABLE IF NOT EXISTS zones (
    region_name             TEXT NOT NULL,
    macrozone_name          TEXT NOT NULL,
    name                    TEXT NOT NULL,
    location                geometry(Polygon, 4326) NOT NULL,
    creation_time           TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (region_name, macrozone_name, name),
    FOREIGN KEY (region_name, macrozone_name)
        REFERENCES macrozones(region_name, name) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_zones_location ON zones USING GIST (location);