-- ========================
-- Regioni
-- ========================
INSERT INTO regions (name, location) VALUES
    ('region-001', ST_SetSRID(ST_GeomFromText('POLYGON((12.45 41.85, 12.55 41.85, 12.55 41.95, 12.45 41.95, 12.45 41.85))'), 4326)),
    ('region-002', ST_SetSRID(ST_GeomFromText('POLYGON((9.10 45.40, 9.30 45.40, 9.30 45.50, 9.10 45.50, 9.10 45.40))'), 4326)),
    ('region-003', ST_SetSRID(ST_GeomFromText('POLYGON((14.20 40.80, 14.40 40.80, 14.40 40.95, 14.20 40.95, 14.20 40.80))'), 4326)),
    ('region-004', ST_SetSRID(ST_GeomFromText('POLYGON((11.20 43.70, 11.40 43.70, 11.40 43.85, 11.20 43.85, 11.20 43.70))'), 4326));

-- ========================
-- Macrozones
-- ========================
INSERT INTO macrozones (region_name, name, location, type, creation_time) VALUES
    ('region-001', 'build-0001', ST_SetSRID(ST_GeomFromText('POLYGON((12.46 41.87, 12.48 41.87, 12.48 41.89, 12.46 41.89, 12.46 41.87))'), 4326), 'indoor', NOW()),
    ('region-001', 'build-0002', ST_SetSRID(ST_GeomFromText('POLYGON((12.50 41.90, 12.52 41.90, 12.52 41.92, 12.50 41.92, 12.50 41.90))'), 4326), 'indoor', NOW()),
    ('region-001', 'macrozone-0003', ST_SetSRID(ST_GeomFromText('POLYGON((12.47 41.91, 12.49 41.91, 12.49 41.93, 12.47 41.93, 12.47 41.91))'), 4326), 'outdoor', NOW()),

    ('region-002', 'build-0004', ST_SetSRID(ST_GeomFromText('POLYGON((9.15 45.42, 9.17 45.42, 9.17 45.44, 9.15 45.44, 9.15 45.42))'), 4326), 'indoor', NOW()),
    ('region-002', 'build-0005', ST_SetSRID(ST_GeomFromText('POLYGON((9.20 45.46, 9.22 45.46, 9.22 45.48, 9.20 45.48, 9.20 45.46))'), 4326), 'indoor', NOW()),
    ('region-002', 'macrozone-0006', ST_SetSRID(ST_GeomFromText('POLYGON((9.25 45.44, 9.27 45.44, 9.27 45.46, 9.25 45.46, 9.25 45.44))'), 4326), 'outdoor', NOW()),

    ('region-003', 'build-0007', ST_SetSRID(ST_GeomFromText('POLYGON((14.25 40.85, 14.27 40.85, 14.27 40.87, 14.25 40.87, 14.25 40.85))'), 4326), 'indoor', NOW()),
    ('region-003', 'macrozone-0008', ST_SetSRID(ST_GeomFromText('POLYGON((14.30 40.90, 14.32 40.90, 14.32 40.92, 14.30 40.92, 14.30 40.90))'), 4326), 'outdoor', NOW()),

    ('region-004', 'build-0009', ST_SetSRID(ST_GeomFromText('POLYGON((11.25 43.75, 11.27 43.75, 11.27 43.77, 11.25 43.77, 11.25 43.75))'), 4326), 'indoor', NOW()),
    ('region-004', 'macrozone-0010', ST_SetSRID(ST_GeomFromText('POLYGON((11.35 43.80, 11.37 43.80, 11.37 43.82, 11.35 43.82, 11.35 43.80))'), 4326), 'outdoor', NOW());

-- ========================
-- Zones
-- ========================
INSERT INTO zones (region_name, macrozone_name, name, location, creation_time) VALUES
    -- Region 1, Build 1
    ('region-001', 'build-0001', 'floor-001', ST_SetSRID(ST_GeomFromText('POLYGON((12.461 41.871, 12.463 41.871, 12.463 41.873, 12.461 41.873, 12.461 41.871))'), 4326), NOW()),
    ('region-001', 'build-0001', 'floor-002', ST_SetSRID(ST_GeomFromText('POLYGON((12.465 41.875, 12.467 41.875, 12.467 41.877, 12.465 41.877, 12.465 41.875))'), 4326), NOW()),

    -- Region 1, Build 2
    ('region-001', 'build-0002', 'floor-001', ST_SetSRID(ST_GeomFromText('POLYGON((12.501 41.901, 12.503 41.901, 12.503 41.903, 12.501 41.903, 12.501 41.901))'), 4326), NOW()),
    ('region-001', 'build-0002', 'floor-002', ST_SetSRID(ST_GeomFromText('POLYGON((12.505 41.905, 12.507 41.905, 12.507 41.907, 12.505 41.907, 12.505 41.905))'), 4326), NOW()),

    -- Region 2, Build 4
    ('region-002', 'build-0004', 'floor-001', ST_SetSRID(ST_GeomFromText('POLYGON((9.151 45.421, 9.153 45.421, 9.153 45.423, 9.151 45.423, 9.151 45.421))'), 4326), NOW()),
    ('region-002', 'build-0004', 'floor-002', ST_SetSRID(ST_GeomFromText('POLYGON((9.155 45.425, 9.157 45.425, 9.157 45.427, 9.155 45.427, 9.155 45.425))'), 4326), NOW()),

    -- Region 2, Build 5
    ('region-002', 'build-0005', 'floor-001', ST_SetSRID(ST_GeomFromText('POLYGON((9.201 45.461, 9.203 45.461, 9.203 45.463, 9.201 45.463, 9.201 45.461))'), 4326), NOW()),
    ('region-002', 'build-0005', 'floor-002', ST_SetSRID(ST_GeomFromText('POLYGON((9.205 45.465, 9.207 45.465, 9.207 45.467, 9.205 45.467, 9.205 45.465))'), 4326), NOW()),

    -- Region 3
    ('region-003', 'build-0007', 'floor-001', ST_SetSRID(ST_GeomFromText('POLYGON((14.251 40.851, 14.253 40.851, 14.253 40.853, 14.251 40.853, 14.251 40.851))'), 4326), NOW()),
    ('region-003', 'build-0007', 'floor-002', ST_SetSRID(ST_GeomFromText('POLYGON((14.255 40.855, 14.257 40.855, 14.257 40.857, 14.255 40.857, 14.255 40.855))'), 4326), NOW()),

    -- Region 4
    ('region-004', 'build-0009', 'floor-001', ST_SetSRID(ST_GeomFromText('POLYGON((11.251 43.751, 11.253 43.751, 11.253 43.753, 11.251 43.753, 11.251 43.751))'), 4326), NOW()),
    ('region-004', 'build-0009', 'floor-002', ST_SetSRID(ST_GeomFromText('POLYGON((11.255 43.755, 11.257 43.755, 11.257 43.757, 11.255 43.757, 11.255 43.755))'), 4326), NOW()),

    -- Zone esterne per zone outdoor esistenti
    ('region-001', 'macrozone-0003', 'zone-001', ST_SetSRID(ST_GeomFromText('POLYGON((12.472 41.912, 12.474 41.912, 12.474 41.914, 12.472 41.914, 12.472 41.912))'), 4326), NOW()),
    ('region-001', 'macrozone-0003', 'zone-002', ST_SetSRID(ST_GeomFromText('POLYGON((12.475 41.915, 12.477 41.915, 12.477 41.917, 12.475 41.917, 12.475 41.915))'), 4326), NOW()),

    ('region-002', 'macrozone-0006', 'zone-001', ST_SetSRID(ST_GeomFromText('POLYGON((9.251 45.445, 9.253 45.445, 9.253 45.447, 9.251 45.447, 9.251 45.445))'), 4326), NOW()),
    ('region-002', 'macrozone-0006', 'zone-002', ST_SetSRID(ST_GeomFromText('POLYGON((9.255 45.448, 9.257 45.448, 9.257 45.450, 9.255 45.450, 9.255 45.448))'), 4326), NOW()),

    ('region-003', 'macrozone-0008', 'zone-001', ST_SetSRID(ST_GeomFromText('POLYGON((14.301 40.901, 14.303 40.901, 14.303 40.903, 14.301 40.903, 14.301 40.901))'), 4326), NOW()),
    ('region-003', 'macrozone-0008', 'zone-002', ST_SetSRID(ST_GeomFromText('POLYGON((14.305 40.905, 14.307 40.905, 14.307 40.907, 14.305 40.907, 14.305 40.905))'), 4326), NOW()),

    ('region-004', 'macrozone-0010', 'zone-001', ST_SetSRID(ST_GeomFromText('POLYGON((11.351 43.801, 11.353 43.801, 11.353 43.803, 11.351 43.803, 11.351 43.801))'), 4326), NOW()),
    ('region-004', 'macrozone-0010', 'zone-002', ST_SetSRID(ST_GeomFromText('POLYGON((11.355 43.805, 11.357 43.805, 11.357 43.807, 11.355 43.807, 11.355 43.805))'), 4326), NOW());