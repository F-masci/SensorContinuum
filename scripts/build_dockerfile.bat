@echo off
setlocal ENABLEDELAYEDEXPANSION

REM === CONTROLLA PARAMETRI: IMMAGINE e VERSIONE ===
if "%~1"=="" (
    echo [USO] build_sensor_agent.bat ^<immagine^> ^<versione^>
    echo Esempio: build_sensor_agent.bat sensor-agent 1.0.0
    exit /b 1
)
if "%~2"=="" (
    echo [ERRORE] Devi specificare anche la versione.
    exit /b 1
)

set "IMAGE_NAME=%~1"
set "VERSION=%~2"

REM === PERCORSO CORRENTE ===
set "CUR_DIR=%cd%"
set "TARGET_DIR=SensorContinuum"

REM === VERIFICA PRESENZA DELLA CARTELLA SENSORCONTINUUM ===
echo %CUR_DIR% | findstr /C:"\%TARGET_DIR%" >nul
if errorlevel 1 (
    echo [ERRORE] La cartella "%TARGET_DIR%" non è nel percorso corrente: %CUR_DIR%
    exit /b 1
)

REM === RISALITA AL ROOT DEL PROGETTO ===
for %%I in ("%CUR_DIR%") do (
    set "FULL_PATH=%%~fI"
)

:find_root
if "%FULL_PATH%"=="" (
    echo [ERRORE] Impossibile trovare la radice del progetto.
    exit /b 1
)

for %%F in ("%FULL_PATH%") do (
    set "FOLDER=%%~nxF"
)

if /I "%FOLDER%"=="%TARGET_DIR%" goto found_root

for %%F in ("%FULL_PATH%") do (
    set "FULL_PATH=%%~dpF"
    set "FULL_PATH=!FULL_PATH:~0,-1!"
)
goto find_root

:found_root
echo [INFO] Trovata radice del progetto: !FULL_PATH!
cd /d "!FULL_PATH!"

REM === COSTRUISCI PATH DEL DOCKERFILE ===
set "DOCKERFILE_PATH=deploy\docker\%IMAGE_NAME%.Dockerfile"

REM === VERIFICA ESISTENZA DOCKERFILE ===
if not exist "!DOCKERFILE_PATH!" (
    echo [ERRORE] Dockerfile non trovato: !DOCKERFILE_PATH!
    exit /b 1
)

REM === BUILD DOCKER ===
echo [INFO] Avvio build Docker con tag: %IMAGE_NAME%:%VERSION%
docker build -f "!DOCKERFILE_PATH!" -t %IMAGE_NAME%:%VERSION% .

if errorlevel 1 (
    echo [ERRORE] La build Docker è fallita.
    exit /b 1
)

echo [OK] Build completata con successo: %IMAGE_NAME%:%VERSION%