#!/bin/bash

# Build script for Coolify Community CLI

set -e

echo "🔨 Building Coolify CLI..."

# Get version info
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT=${COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo "unknown")}
DATE=${DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}

# Build for multiple platforms
PLATFORMS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"

for platform in $PLATFORMS; do
    GOOS=$(echo $platform | cut -d'/' -f1)
    GOARCH=$(echo $platform | cut -d'/' -f2)
    
    OUTPUT="dist/coolify-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi
    
    echo "Building for $GOOS/$GOARCH..."
    
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
        -o $OUTPUT \
        .
done

echo "✅ Build completed!"
echo "Binaries are available in the dist/ directory"
