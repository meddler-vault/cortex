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

# Define a build argument for the version
ARG WATCHDOG_VERSION
ENV WATCHDOG_VERSION=${WATCHDOG_VERSION}
ENV message_queue_topic=tasks_publish



RUN CGO_ENABLED=0  go build -ldflags "-X github.com/meddler-vault/cortex/consumer-nats.WatchdogVersion=$WATCHDOG_VERSION" -o /opt/watchdog && echo "Build complete. Contents of /opt:" && ls -l /opt/


RUN echo "Built: Watchdog version: $WATCHDOG_VERSION"

RUN chmod +x /opt/watchdog


FROM bash

COPY --from=build /opt/watchdog /opt/watchdog



RUN echo "Build complete. Contents of /opt:" && ls -l /opt/







