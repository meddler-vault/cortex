
docker build  --progress=plain --no-cache  --build-arg WATCHDOG_VERSION=$(date +%Y.%m.%d.%H.%M.%S)  -t rounak316/watchcat:linux .

docker tag rounak316/watchcat:linux rounak316/watchcat:linux
docker push rounak316/watchcat:linux



