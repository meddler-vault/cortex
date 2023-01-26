
TZ=Asia/Calcutta 

GOOS=linux GOARCH=amd64  go build -ldflags "-X github.com/meddler-io/watchdog/consumer.WatchdogVersion=`date +%Y.%m.%d.%H.%M.%S`"  -o ./artifacts/watchdog 
cd ./artifacts
./build.sh