# Set timezone and build the Go project
TZ=Asia/Calcutta

# Assign the version to a variable
WATCHDOG_VERSION=$(date +%Y.%m.%d.%H.%M.%S)

# Build the Go project with the specified OS and architecture settings
#GOOS=linux GOARCH=amd64 
go build -ldflags "-X github.com/meddler-vault/cortex/consumer-nats.WatchdogVersion=$WATCHDOG_VERSION" -o ./artifacts/watchdog

# Echo the watchdog version
echo "Watchdog version: $WATCHDOG_VERSION"

# Change to the artifacts directory
pushd ./artifacts

# Execute the build script
./build.sh

# Return to the original directory
popd
