-- Regioni
INSERT INTO regions (name) VALUES
    ('region-001'),
    ('region-002');

-- Macrozone (Edifici e Aree) con posizione geografica
INSERT INTO macrozones (region_name, name, location, creation_time) VALUES
    ('region-001', 'build-0001', ST_SetSRID(ST_MakePoint(12.4924, 41.8902), 4326), NOW()),
    ('region-002', 'build-0002', ST_SetSRID(ST_MakePoint(12.4964, 41.9028), 4326), NOW()),
    ('region-001', 'macrozone-0003', ST_SetSRID(ST_MakePoint(9.1900, 45.4642), 4326), NOW()),
    ('region-002', 'macrozone-0004', ST_SetSRID(ST_MakePoint(9.1859, 45.4654), 4326), NOW());

-- Zone all'interno delle macrozone
INSERT INTO zones (region_name, macrozone_name, name, creation_time) VALUES
    ('region-001', 'build-0001', 'floor-001', NOW()),
    ('region-001', 'build-0001', 'floor-002', NOW()),
    ('region-002', 'build-0002', 'floor-001', NOW()),
    ('region-002', 'build-0002', 'floor-002', NOW()),
    ('region-001', 'macrozone-0003', 'zone-001', NOW()),
    ('region-001', 'macrozone-0003', 'zone-002', NOW()),
    ('region-002', 'macrozone-0004', 'zone-001', NOW()),
    ('region-002', 'macrozone-0004', 'zone-002', NOW());