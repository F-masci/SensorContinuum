-- Abilita estensioni necessarie
CREATE EXTENSION IF NOT EXISTS postgis;

-- 1. Tabella regioni
CREATE TABLE IF NOT EXISTS regions (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);