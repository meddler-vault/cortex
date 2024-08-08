sudo git pull


export PATH=$PATH:$(go env GOPATH)/bin



docker build  --progress=plain --no-cache  --build-arg WATCHDOG_VERSION=$(date +%Y.%m.%d.%H.%M.%S)  -t rounak316/watchdog:linux .

docker tag rounak316/watchdog:linux rounak316/watchdog:linux
docker push rounak316/watchdog:linux



