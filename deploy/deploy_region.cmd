set STACK_NAME=RegionStack_001
set TEMPLATE_FILE=..\deploy\terraform\region.yaml
set INSTANCE_TYPE=t3.medium
set ENDPOINT=http://localhost:4566

aws --endpoint-url=%ENDPOINT% cloudformation update-stack ^
    --stack-name %STACK_NAME% ^
    --template-body file://%TEMPLATE_FILE% ^
    --parameters ParameterKey=InstanceType,ParameterValue=%INSTANCE_TYPE%