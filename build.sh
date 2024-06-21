# Set timezone and build the Go project
TZ=Asia/Calcutta

# Build the Go project with specific OS and architecture settings
GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/meddler-io/watchdog/consumer.WatchdogVersion=`date +%Y.%m.%d.%H.%M.%S`" -o ./artifacts/watchdog 

# Change to the artifacts directory
pushd ./artifacts

# Execute the build script
./build.sh

# Return to the original directory
popd
