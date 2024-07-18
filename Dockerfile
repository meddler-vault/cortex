# Use the official Golang image as the build stage
FROM golang:latest AS build

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN TZ=Asia/Calcutta
RUN WATCHDOG_VERSION=$(date +%Y.%m.%d.%H.%M.%S)
RUN echo "Building: Watchdog version: $WATCHDOG_VERSION"


RUN go build -ldflags "-X github.com/meddler-vault/cortex/consumer-nats.WatchdogVersion=$WATCHDOG_VERSION" -o /opt/watchdog


RUN echo "Built: Watchdog version: $WATCHDOG_VERSION"


FROM scratch

COPY --from=build /opt/watchdog /go/src/github.com/meddler-vault/cortex/watchdog.bin




