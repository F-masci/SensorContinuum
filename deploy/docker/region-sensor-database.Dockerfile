FROM timescale/timescaledb:latest-pg17

COPY configs/postgresql/init-region-sensors-db.sql /docker-entrypoint-initdb.d/init-region-sensors-db.sql