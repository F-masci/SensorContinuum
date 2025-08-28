-- Tabella degli hub della regione (intermediate fog hub)
CREATE TABLE IF NOT EXISTS region_hubs (
    id                  TEXT PRIMARY KEY,           -- UUID o identificativo univoco
    service             TEXT NOT NULL,              -- tipo di servizio/ruolo
    registration_time   TIMESTAMP,
    last_seen           TIMESTAMP
);

-- Tabella degli hub di macrozona (proximity fog hub)
CREATE TABLE IF NOT EXISTS macrozone_hubs (
    id                  TEXT NOT NULL,
    macrozone_name      TEXT NOT NULL,
    service             TEXT NOT NULL,
    registration_time   TIMESTAMP,
    last_seen           TIMESTAMP,
    PRIMARY KEY (id, macrozone_name)
);

-- Tabella degli hub di zona (edge hub)
CREATE TABLE IF NOT EXISTS zone_hubs (
    id                  TEXT NOT NULL,
    macrozone_name      TEXT NOT NULL,
    zone_name           TEXT NOT NULL,
    service             TEXT NOT NULL,
    registration_time   TIMESTAMP,
    last_seen           TIMESTAMP,
    PRIMARY KEY (id, macrozone_name, zone_name)
);

-- Tabella dei sensori associati agli edge hub (anche se non sono direttamente collegati a un hub specifico)
CREATE TABLE IF NOT EXISTS sensors (
    id                  TEXT NOT NULL,
    macrozone_name      TEXT NOT NULL,
    zone_name           TEXT NOT NULL,
    type                TEXT,
    reference           TEXT,
    registration_time   TIMESTAMP,
    last_seen           TIMESTAMP,
    PRIMARY KEY (id, macrozone_name, zone_name)
);