#!/bin/bash

# Build script for mcp-cli-go
# Builds binaries for all supported platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Version information
VERSION="${VERSION:-dev}"
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build directory
BUILD_DIR="dist"

echo -e "${GREEN}Building mcp-cli-go${NC}"
echo "Version: $VERSION"
echo "Build Time: $BUILD_TIME"
echo "Git Commit: $GIT_COMMIT"
echo ""

# Clean build directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Build flags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X 'main.Version=${VERSION}'"
LDFLAGS="$LDFLAGS -X 'main.BuildTime=${BUILD_TIME}'"
LDFLAGS="$LDFLAGS -X 'main.GitCommit=${GIT_COMMIT}'"
LDFLAGS="$LDFLAGS -extldflags '-static'"

# Build tags for static linking
BUILD_TAGS="netgo"

# Array of platforms to build
declare -a platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# Build for each platform
for platform in "${platforms[@]}"; do
    IFS='/' read -r -a parts <<< "$platform"
    GOOS="${parts[0]}"
    GOARCH="${parts[1]}"
    
    output_name="mcp-cli-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo -e "${YELLOW}Building for ${GOOS}/${GOARCH}...${NC}"
    
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build \
        -a \
        -tags "$BUILD_TAGS" \
        -installsuffix netgo \
        -ldflags="$LDFLAGS" \
        -o "${BUILD_DIR}/${output_name}" \
        .
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Built ${output_name}${NC}"
        
        # Create checksum
        cd "$BUILD_DIR"
        sha256sum "${output_name}" > "${output_name}.sha256"
        cd ..
    else
        echo -e "${RED}✗ Failed to build ${output_name}${NC}"
        exit 1
    fi
done

echo ""
echo -e "${GREEN}Build complete!${NC}"
echo ""
echo "Binaries created in ${BUILD_DIR}:"
ls -lh "$BUILD_DIR"

echo ""
echo -e "${YELLOW}Testing local binary...${NC}"
chmod +x "${BUILD_DIR}/mcp-cli-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')"
"${BUILD_DIR}/mcp-cli-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')" --version

echo ""
echo -e "${GREEN}All builds successful!${NC}"
