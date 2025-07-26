@echo off
setlocal

REM Prende il nome della funzione come parametro (opzionale)
set LAMBDA_NAME=%1

cd ..

if "%LAMBDA_NAME%"=="" (
    docker build -t golambda . -f deploy/docker/lambda.Dockerfile
) else (
    docker build -t golambda . -f deploy/docker/lambda.Dockerfile --build-arg LAMBDA_PATH=%LAMBDA_NAME%
)

docker create --name tmpcontainer golambda

REM Crea una cartella temporanea per estrarre tutti i file
mkdir lambda_tmp
docker cp tmpcontainer:/lambda/. ./lambda_tmp

REM Cancella la cartella di destinazione se esiste e ricreala
rmdir /s /q lambda
mkdir lambda

REM Copia solo i file .zip mantenendo la struttura delle cartelle
xcopy /s /e /y /i lambda_tmp\*.zip lambda\

REM Pulisci la cartella temporanea
rmdir /s /q lambda_tmp

docker rm tmpcontainer
cd scripts