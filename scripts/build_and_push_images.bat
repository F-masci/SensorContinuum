call .\scripts\build_dockerfile.bat deploy\docker\sensor-agent.Dockerfile fmasci/sc-sensor-agent latest
call .\scripts\build_dockerfile.bat deploy\docker\edge-hub.Dockerfile fmasci/sc-edge-hub latest
call .\scripts\build_dockerfile.bat deploy\docker\proximity-fog-hub.Dockerfile fmasci/sc-proximity-fog-hub latest
call .\scripts\build_dockerfile.bat deploy\docker\intermediate-fog-hub.Dockerfile fmasci/sc-intermediate-fog-hub latest

docker push fmasci/sc-sensor-agent:latest
docker push fmasci/sc-edge-hub:latest
docker push fmasci/sc-proximity-fog-hub:latest
docker push fmasci/sc-intermediate-fog-hub:latest