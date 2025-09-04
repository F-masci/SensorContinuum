FROM timescale/timescaledb:latest-pg17

COPY configs/postgresql/init-proximity-cache.sql /docker-entrypoint-initdb.d/init-proximity-cache.sql