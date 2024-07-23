
docker build --build-arg WATCHDOG_VERSION=$(date +%Y.%m.%d.%H.%M.%S)  --no-cache -t rounak316/watchdog:linux .

docker tag rounak316/watchdog:linux rounak316/watchdog:linux
docker push rounak316/watchdog:linux



