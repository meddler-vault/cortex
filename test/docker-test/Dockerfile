# Use an intermediate image with useful tools
FROM alpine:latest AS builder

# Install required packages
RUN apk add --no-cache curl jq

RUN set -x && echo "hello-world"

# Get platform and architecture information
RUN \
    PLATFORM=$(uname -s) && \
    ARCH=$(uname -m) && \
    echo "Platform: $PLATFORM" && \
    echo "Architecture: $ARCH" && \
    echo "PLATFORM=$PLATFORM" > /platform_info.txt && \
    echo "ARCH=$ARCH" >> /platform_info.txt


# Set the GitHub repository and binary details
ENV REPO_OWNER=meddler-vault
ENV REPO_NAME=cortex
ENV BIN_NAME=platform-arch

# Define the GitHub API URL for latest release
ENV API_URL=https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest

# Get the latest release tag and download the binary
RUN \
    # Get the latest release tag
    RELEASE_TAG=$(curl -s ${API_URL} | jq -r .tag_name) && \
    echo "Latest release tag: $RELEASE_TAG" && \
    \
    # Construct the download URL for the binary
    DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${RELEASE_TAG}/${BIN_NAME}-${RELEASE_TAG}-${ARCH}-${PLATFORM}.tar.gz" && \
    echo "Download URL: $DOWNLOAD_URL" && \
    \
    # Download and extract the binary
    curl -L ${DOWNLOAD_URL} | tar xz -C /tmp && \
    \
    # Rename the binary
    mv /tmp/${BIN_NAME}-${RELEASE_TAG}-${ARCH}-${PLATFORM} /tmp/${BIN_NAME} && \
    \
    # Copy the binary to the final image
    cp /tmp/${BIN_NAME} /go/bin/${BIN_NAME}

# Create the final scratch image
FROM scratch

# Copy the binary from the builder stage
COPY --from=builder /go/bin/platform-arch /platform-arch

# Set the entrypoint
ENTRYPOINT ["/platform-arch"]
