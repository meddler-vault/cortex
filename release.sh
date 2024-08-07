#!/bin/bash

# GitHub repository details
REPO="meddler-vault/cortex"
RELEASE_TAG="v0.1"  # Change this to your release version
RELEASE_NAME="Release $RELEASE_TAG"
RELEASE_BODY="This is the release description for $RELEASE_TAG"

# Create a new release on GitHub
echo "Creating release $RELEASE_TAG..."
gh release create "$RELEASE_TAG" --title "$RELEASE_NAME" --notes "$RELEASE_BODY" --repo "$REPO"

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

# Directory where binaries are stored
output_dir="build_for_all_test/"

# Loop through each platform and upload the binaries
for platform in "${platforms[@]}"
do
    IFS="/" read -r -a arr <<< "$platform"
    GOOS="${arr[0]}"
    GOARCH="${arr[1]}"
    output_file="${output_dir}/${GOOS}-${GOARCH}"

    if [ -f "$output_file" ]; then
        echo "Uploading $output_file..."
        gh release upload "$RELEASE_TAG" "$output_file" --repo "$REPO"
    else
        echo "File $output_file not found. Skipping upload."
    fi
done

echo "Upload completed."
