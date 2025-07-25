cd ../../../
docker build -t golambda . -f cmd/api-backend/region/regionList.Dockerfile
cd cmd/api-backend/region/
docker create --name tmpcontainer golambda
docker cp tmpcontainer:/lambda/regionList.zip ./regionList.zip
docker rm tmpcontainer