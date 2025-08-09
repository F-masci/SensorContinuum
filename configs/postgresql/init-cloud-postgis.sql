-- 1. Tabella regioni con name come chiave primaria
CREATE TABLE IF NOT EXISTS regions (
    name                TEXT PRIMARY KEY
);

-- 2. Tabella macrozone con chiave primaria composta
CREATE TABLE IF NOT EXISTS macrozones (
    region_name         TEXT NOT NULL REFERENCES regions(name) ON DELETE CASCADE,
    name                TEXT NOT NULL,
    location            geometry(Point, 4326) NOT NULL,
    creation_time       TIMESTAMP,
    PRIMARY KEY (region_name, name)
);

CREATE INDEX IF NOT EXISTS idx_macrozones_location ON macrozones USING GIST (location);

-- 3. Tabella zone con chiave primaria composta
CREATE TABLE IF NOT EXISTS zones (
    region_name             TEXT NOT NULL,
    macrozone_name          TEXT NOT NULL,
    name                    TEXT NOT NULL,
    creation_time           TIMESTAMP,
    PRIMARY KEY (region_name, macrozone_name, name),
    FOREIGN KEY (region_name, macrozone_name)
        REFERENCES macrozones(region_name, name) ON DELETE CASCADE
);