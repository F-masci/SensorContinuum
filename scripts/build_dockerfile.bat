@echo off
setlocal ENABLEDELAYEDEXPANSION

REM === CONTROLLA PARAMETRI: DOCKERFILE, IMMAGINE e VERSIONE ===
if "%~1"=="" (
    echo [USO] build_dockerfile.bat ^<dockerfile-path^> ^<immagine^> ^<versione^>
    echo Esempio: build_dockerfile.bat deploy\docker\sensor-agent.Dockerfile sensor-agent 1.0.0
    exit /b 1
)
if "%~2"=="" (
    echo [ERRORE] Devi specificare anche il nome dell'immagine.
    exit /b 1
)
if "%~3"=="" (
    echo [ERRORE] Devi specificare anche la versione.
    exit /b 1
)

set "DOCKERFILE_PATH=%~1"
set "IMAGE_NAME=%~2"
set "VERSION=%~3"

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