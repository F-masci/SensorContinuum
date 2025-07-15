@echo off
setlocal ENABLEDELAYEDEXPANSION

REM === CONTROLLA PARAMETRO VERSIONE ===
if "%~1"=="" (
    echo [USO] build_sensor_agent.bat ^<versione^>
    echo Esempio: build_sensor_agent.bat 1.0.0
    exit /b 1
)
set "VERSION=%~1"

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

REM === BUILD DOCKER ===
echo [INFO] Avvio build Docker con tag: sensor-agent:%VERSION%
docker build -f deploy\docker\sensor-agent.Dockerfile -t sensor-agent:%VERSION% .

if errorlevel 1 (
    echo [ERRORE] La build Docker è fallita.
    exit /b 1
)

echo [OK] Build completata con successo: sensor-agent:%VERSION%