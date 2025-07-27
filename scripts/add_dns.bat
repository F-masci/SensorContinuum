@echo off
setlocal

if "%~1"=="" (
  echo Utilizzo: %~nx0 HOSTNAME IP
  exit /b 1
)
if "%~2"=="" (
  echo Utilizzo: %~nx0 HOSTNAME IP
  exit /b 1
)

set RECORD_NAME=%~1
set IP=%~2
set ENDPOINT=http://localhost:4566
set ZONE_NAME=sensorcontinuum.node
set JSON_FILE=%TEMP%\dns_change_%RANDOM%.json
set ZONE_ID=

REM Trova l'ID della hosted zone
for /f "delims=" %%i in ('aws route53 list-hosted-zones --endpoint-url=%ENDPOINT% --query "HostedZones[?Name=='%ZONE_NAME%.'].Id" --output text') do (
  set ZONE_ID=%%i
)

REM Se non esiste, crea la hosted zone
if "%ZONE_ID%"=="" (
  for /f "delims=" %%i in ('aws route53 create-hosted-zone --name %ZONE_NAME% --caller-reference %RANDOM% --endpoint-url=%ENDPOINT% --query "HostedZone.Id" --output text') do (
    set ZONE_ID=%%i
  )
)

REM Rimuovi eventuale prefisso "/hostedzone/"
set ZONE_ID=%ZONE_ID:/hostedzone/=%
if "%ZONE_ID%"=="" (
  echo Errore: impossibile ottenere l'ID della hosted zone.
  exit /b 1
)

REM Crea il file JSON temporaneo riga per riga
> "%JSON_FILE%" echo {
>> "%JSON_FILE%" echo   "Comment": "Aggiorna record A",
>> "%JSON_FILE%" echo   "Changes": [
>> "%JSON_FILE%" echo     {
>> "%JSON_FILE%" echo       "Action": "UPSERT",
>> "%JSON_FILE%" echo       "ResourceRecordSet": {
>> "%JSON_FILE%" echo         "Name": "%RECORD_NAME%",
>> "%JSON_FILE%" echo         "Type": "A",
>> "%JSON_FILE%" echo         "TTL": 300,
>> "%JSON_FILE%" echo         "ResourceRecords": [
>> "%JSON_FILE%" echo           {
>> "%JSON_FILE%" echo             "Value": "%IP%"
>> "%JSON_FILE%" echo           }
>> "%JSON_FILE%" echo         ]
>> "%JSON_FILE%" echo       }
>> "%JSON_FILE%" echo     }
>> "%JSON_FILE%" echo   ]
>> "%JSON_FILE%" echo }

aws route53 change-resource-record-sets --endpoint-url=%ENDPOINT% --hosted-zone-id %ZONE_ID% --change-batch file://%JSON_FILE%

del "%JSON_FILE%"

echo Record DNS aggiornato!
endlocal