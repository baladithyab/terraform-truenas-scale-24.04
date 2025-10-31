#!/bin/bash
set -e

# Build script for TrueNAS Terraform Provider releases
# Usage: ./build-release.sh <version>
# Example: ./build-release.sh v0.2.13

VERSION=${1:-"dev"}
BINARY_NAME="terraform-provider-truenas"
DIST_DIR="dist"

# Remove 'v' prefix if present for directory naming
VERSION_NUM=${VERSION#v}

# Platforms to build for
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

echo "Building ${BINARY_NAME} ${VERSION} for multiple platforms..."

# Clean previous builds
rm -rf ${DIST_DIR}
mkdir -p ${DIST_DIR}

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT_NAME="${BINARY_NAME}_${VERSION}_${GOOS}_${GOARCH}"
    
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    echo "Building for ${GOOS}/${GOARCH}..."
    GOOS=$GOOS GOARCH=$GOARCH go build -o "${DIST_DIR}/${OUTPUT_NAME}" .
    
    if [ $? -ne 0 ]; then
        echo "Error building for ${GOOS}/${GOARCH}"
        exit 1
    fi
done

# Generate SHA256 checksums
echo "Generating SHA256 checksums..."
cd ${DIST_DIR}
sha256sum ${BINARY_NAME}_${VERSION}_* > ${BINARY_NAME}_${VERSION}_SHA256SUMS
cd ..

echo ""
echo "Build complete! Binaries are in ${DIST_DIR}/"
echo ""
echo "Files created:"
ls -lh ${DIST_DIR}/

echo ""
echo "Next steps:"
echo "1. Create a GitHub release for ${VERSION}"
echo "2. Upload all files from ${DIST_DIR}/ to the release"
echo "3. Use RELEASE_NOTES_${VERSION}.md as the release description"

