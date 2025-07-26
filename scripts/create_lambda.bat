@echo off
setlocal enabledelayedexpansion

REM Parametro opzionale: nome funzione lambda (es: region/regionList)
set ENDPOINT=http://localhost:4566
set LAMBDA_NAME=%1
set ROLE_ARN=arn:aws:iam::000000000000:role/fake-role
set REGION=us-east-1

cd ..

if "%LAMBDA_NAME%"=="" (
    dir /b /s lambda\*.zip >nul 2>&1
    if errorlevel 1 (
        echo Nessun file .zip trovato nella cartella lambda.
    ) else (
        for /r lambda %%Z in (*.zip) do (
            set "FUNC=%%~nZ"
            echo Eliminazione, se esiste, funzione !FUNC!...
            aws lambda delete-function ^
                --endpoint-url=%ENDPOINT% ^
                --function-name !FUNC! ^
                --region %REGION% >nul 2>&1
            echo Creazione funzione !FUNC!...
            aws lambda create-function ^
                --endpoint-url=%ENDPOINT% ^
                --function-name !FUNC! ^
                --runtime go1.x ^
                --handler !FUNC! ^
                --zip-file fileb://%%Z ^
                --role %ROLE_ARN% ^
                --region %REGION%
        )
    )
) else (
    if exist lambda\%LAMBDA_NAME%.zip (
        set "FUNC_NAME=%LAMBDA_NAME%"
        for %%A in ("!FUNC_NAME!") do set "FUNC_NAME=%%~nxA"
        echo Eliminazione, se esiste, funzione !FUNC_NAME!...
        aws lambda delete-function ^
            --endpoint-url=%ENDPOINT% ^
            --function-name !FUNC_NAME! ^
            --region %REGION% >nul 2>&1
        echo Creazione funzione !FUNC_NAME!...
        aws lambda create-function ^
            --endpoint-url=%ENDPOINT% ^
            --function-name !FUNC_NAME! ^
            --runtime go1.x ^
            --handler !FUNC_NAME! ^
            --zip-file fileb://lambda\%LAMBDA_NAME%.zip ^
            --role %ROLE_ARN% ^
            --region %REGION%
    ) else (
        echo La funzione %LAMBDA_NAME% non esiste come file lambda\%LAMBDA_NAME%.zip.
    )
)

cd scripts
pause