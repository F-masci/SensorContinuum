@echo off
setlocal enabledelayedexpansion

set ENDPOINT=http://localhost:4566
set REGION=us-east-1
set ROLE_ARN=arn:aws:iam::000000000000:role/fake-role
set STAGE_NAME=dev

cd ..

echo. > lambda/endpoints.txt

REM Cicla tutti i file .zip nella cartella lambda e sottocartelle
for /r lambda %%F in (*.zip) do (
    set "FUNC=%%~nF"
    for %%D in ("%%~dpF.") do (
        set "SUBFOLDER=%%~nxD"
    )

    REM Controlla se abbiamo già creato l'API per questa sottocartella
    call set "API_ID=%%API_ID_!SUBFOLDER!%%"
    if "!API_ID!"=="" (
        for /f "tokens=*" %%i in ('aws --endpoint-url=%ENDPOINT% apigateway create-rest-api --name "!SUBFOLDER!-api" --query "id" --output text') do set API_ID=%%i
        set "API_ID_!SUBFOLDER!=!API_ID!"
        for /f "tokens=*" %%i in ('aws --endpoint-url=%ENDPOINT% apigateway get-resources --rest-api-id !API_ID! --query "items[?path==`/`].id" --output text') do set PARENT_ID=%%i
        for /f "tokens=*" %%i in ('aws --endpoint-url=%ENDPOINT% apigateway create-resource --rest-api-id !API_ID! --parent-id !PARENT_ID! --path-part !SUBFOLDER! --query "id" --output text') do set SUBFOLDER_ID=%%i
        set "SUBFOLDER_ID_!SUBFOLDER!=!SUBFOLDER_ID!"
    ) else (
        call set "API_ID=%%API_ID_!SUBFOLDER!%%"
        call set "SUBFOLDER_ID=%%SUBFOLDER_ID_!SUBFOLDER!%%"
    )

    REM Mappa nome funzione in path
    set "PATH_PART="
    if /i "!FUNC!"=="regionList" (
        set "PATH_PART=list"
    ) else if /i "!FUNC!"=="regionDetail" (
        set "PATH_PART=detail"
        set "HAS_ID=1"
    ) else (
        set "PATH_PART=!FUNC!"
    )

    REM Crea la risorsa secondaria (es: /region/list o /region/detail/{id})
    if defined HAS_ID (
        for /f "tokens=*" %%i in ('aws --endpoint-url=%ENDPOINT% apigateway create-resource --rest-api-id !API_ID! --parent-id !SUBFOLDER_ID! --path-part !PATH_PART! --query "id" --output text') do set RESOURCE_ID=%%i
        for /f "tokens=*" %%i in ('aws --endpoint-url=%ENDPOINT% apigateway create-resource --rest-api-id !API_ID! --parent-id !RESOURCE_ID! --path-part "{id}" --query "id" --output text') do set RESOURCE_ID=%%i
        set "RESOURCE_PATH=!SUBFOLDER!/!PATH_PART!/{id}"
        set "HAS_ID="
    ) else (
        for /f "tokens=*" %%i in ('aws --endpoint-url=%ENDPOINT% apigateway create-resource --rest-api-id !API_ID! --parent-id !SUBFOLDER_ID! --path-part !PATH_PART! --query "id" --output text') do set RESOURCE_ID=%%i
        set "RESOURCE_PATH=!SUBFOLDER!/!PATH_PART!"
    )

    aws --endpoint-url=%ENDPOINT% apigateway put-method ^
        --rest-api-id !API_ID! ^
        --resource-id !RESOURCE_ID! ^
        --http-method GET ^
        --authorization-type NONE >nul

    aws --endpoint-url=%ENDPOINT% apigateway put-integration ^
        --rest-api-id !API_ID! ^
        --resource-id !RESOURCE_ID! ^
        --http-method GET ^
        --type AWS_PROXY ^
        --integration-http-method POST ^
        --uri arn:aws:apigateway:%REGION%:lambda:path/2015-03-31/functions/arn:aws:lambda:%REGION%:000000000000:function:!FUNC!/invocations >nul

    aws --endpoint-url=%ENDPOINT% lambda remove-permission ^
        --function-name !FUNC! ^
        --statement-id apigateway-test-!FUNC! >nul 2>&1

    aws --endpoint-url=%ENDPOINT% lambda add-permission ^
        --function-name !FUNC! ^
        --statement-id apigateway-test-!FUNC! ^
        --action lambda:InvokeFunction ^
        --principal apigateway.amazonaws.com ^
        --source-arn arn:aws:execute-api:%REGION%:000000000000:!API_ID!/*/GET/!RESOURCE_PATH! >nul

    aws --endpoint-url=%ENDPOINT% apigateway create-deployment ^
        --rest-api-id !API_ID! ^
        --stage-name %STAGE_NAME% >nul

    set URL=%ENDPOINT%/restapis/!API_ID!/%STAGE_NAME%/_user_request_/!RESOURCE_PATH!
    echo !FUNC! !URL!>> lambda/endpoints.txt
)

echo Endpoint creati:
type lambda/endpoints.txt

cd scripts
pause
exit /b