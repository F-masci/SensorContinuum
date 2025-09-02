@echo off
setlocal

set ENDPOINT=http://localhost:4566
set STACK_NAME=CloudStack-v0-0-2
set TEMPLATE_FILE=..\deploy\terraform\cloud.yaml

aws cloudformation create-stack ^
    --stack-name %STACK_NAME% ^
    --template-body file://%TEMPLATE_FILE% ^
    --parameters ParameterKey=InitSQLBucket,ParameterValue=my-init-scripts-bucket ^
                 ParameterKey=InitSQLKey,ParameterValue=init-cloud-metadata-db.sql ^
                 ParameterKey=DBMasterPassword,ParameterValue=adminpass ^
    --capabilities CAPABILITY_NAMED_IAM

endlocal