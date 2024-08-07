#!/bin/bash

# List of OS/Arch combinations to build
platforms=(

    "linux/amd64"
    "linux/386"
    "linux/arm"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "freebsd/amd64"
    "freebsd/386"
    "freebsd/arm"
)

# Name of the output binary
output_name="myapp"

# Loop through each platform and build
for platform in "${platforms[@]}"
do
    IFS="/" read -r -a arr <<< "$platform"
    GOOS="${arr[0]}"
    GOARCH="${arr[1]}"
    output_dir="build_for_all_test/"
    output_file="${output_dir}/${GOOS}-${GOARCH}"

    # Create the output directory if it doesn't exist
    mkdir -p "${output_dir}"

    # Set the environment variables and build the binary
    WATCHDOG_VERSION=0.1
    env GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "-X github.com/meddler-vault/cortex/consumer-nats.WatchdogVersion=$WATCHDOG_VERSION" -o ${output_file} ;
    # env GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${output_file} main.go


    if [ $? -ne 0 ]; then
        echo "An error occurred while building for ${platform}"
        exit 1
    fi

    echo "Built ${output_file}"
done

echo "Cross-compilation completed."
