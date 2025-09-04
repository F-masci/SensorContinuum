FROM postgis/postgis:latest

COPY configs/postgresql/init-region-metadata-db.sql /docker-entrypoint-initdb.d/init-region-metadata-db.sql