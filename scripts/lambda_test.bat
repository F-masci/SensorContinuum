@echo off
setlocal enabledelayedexpansion

set ENDPOINT=http://localhost:4566
set REGION=us-east-1
set LAMBDA_NAME=regionList
set ROLE_ARN=arn:aws:iam::000000000000:role/fake-role
set API_NAME=region-api
set STAGE_NAME=dev
set RESOURCE_PATH=region

echo Creazione API Gateway...

for /f "tokens=*" %%i in ('aws --endpoint-url=%ENDPOINT% apigateway create-rest-api --name "%API_NAME%" --query "id" --output text') do set API_ID=%%i

echo API creata: %API_ID%

for /f "tokens=*" %%i in ('aws --endpoint-url=%ENDPOINT% apigateway get-resources --rest-api-id %API_ID% --query "items[?path=='/'].id" --output text') do set PARENT_ID=%%i

echo Root resource ID: %PARENT_ID%

for /f "tokens=*" %%i in ('aws --endpoint-url=%ENDPOINT% apigateway create-resource --rest-api-id %API_ID% --parent-id %PARENT_ID% --path-part %RESOURCE_PATH% --query "id" --output text') do set RESOURCE_ID=%%i

echo Risorsa creata: %RESOURCE_ID%

echo Aggiunta metodo GET...

aws --endpoint-url=%ENDPOINT% apigateway put-method ^
    --rest-api-id %API_ID% ^
    --resource-id %RESOURCE_ID% ^
    --http-method GET ^
    --authorization-type NONE >nul

echo Collegamento Lambda con metodo GET...

aws --endpoint-url=%ENDPOINT% apigateway put-integration ^
    --rest-api-id %API_ID% ^
    --resource-id %RESOURCE_ID% ^
    --http-method GET ^
    --type AWS_PROXY ^
    --integration-http-method POST ^
    --uri arn:aws:apigateway:%REGION%:lambda:path/2015-03-31/functions/arn:aws:lambda:%REGION%:000000000000:function:%LAMBDA_NAME%/invocations >nul

echo Permessi Lambda per invocazione da API Gateway...

aws --endpoint-url=%ENDPOINT% lambda add-permission ^
    --function-name %LAMBDA_NAME% ^
    --statement-id apigateway-test-2 ^
    --action lambda:InvokeFunction ^
    --principal apigateway.amazonaws.com ^
    --source-arn arn:aws:execute-api:%REGION%:000000000000:%API_ID%/*/GET/%RESOURCE_PATH% >nul

echo Deploy API Gateway...

aws --endpoint-url=%ENDPOINT% apigateway create-deployment ^
    --rest-api-id %API_ID% ^
    --stage-name %STAGE_NAME% >nul

echo API Gateway pubblicata.

pause