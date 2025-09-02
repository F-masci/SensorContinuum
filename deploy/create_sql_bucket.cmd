@echo off
setlocal

REM Endpoint di LocalStack
set ENDPOINT=http://localhost:4566
set BUCKET_NAME=my-init-scripts-bucket

REM Verifica se il bucket esiste
aws s3api head-bucket --bucket %BUCKET_NAME% >nul 2>&1
if %ERRORLEVEL%==0 (
    echo Bucket %BUCKET_NAME% già esistente
) else (
    echo Creo bucket %BUCKET_NAME%
    aws s3api create-bucket --bucket %BUCKET_NAME%
)

REM Caricamento di un file SQL di esempio
set SQL_FILE_NAME=init-cloud-metadata-db.sql
set SQL_FILE=../configs/postgresql/init-cloud-metadata-db.sql
if exist %SQL_FILE% (
    echo Carico %SQL_FILE% su %BUCKET_NAME%
    aws s3 cp %SQL_FILE% s3://%BUCKET_NAME%/%SQL_FILE_NAME%
) else (
    echo File %SQL_FILE% non trovato
)

endlocal
