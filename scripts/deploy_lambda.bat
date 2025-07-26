@echo off
setlocal

echo === Zip delle funzioni Lambda ===
call zip_lambda.bat
if errorlevel 1 (
    echo Errore durante la zip delle funzioni Lambda.
    exit /b 1
)

echo === Deploy delle funzioni Lambda ===
call create_lambda.bat
if errorlevel 1 (
    echo Errore durante la creazione delle funzioni Lambda.
    exit /b 1
)

echo === Creazione degli endpoint API Gateway ===
REM call create_endpoint.bat
REM if errorlevel 1 (
REM     echo Errore durante la creazione degli endpoint.
REM     exit /b 1
REM )

echo === Deploy completato con successo! ===
pause