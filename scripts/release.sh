#!/bin/bash

# Release script for Coolify Community CLI

set -e

echo "🚀 Creating Coolify CLI release..."

# Check if we're on a tag
if ! git describe --tags --exact-match >/dev/null 2>&1; then
    echo "❌ Not on a git tag. Create a tag first."
    exit 1
fi

VERSION=$(git describe --tags --exact-match)
echo "Releasing version: $VERSION"

# Build
./scripts/build.sh

# Create release archive
cd dist
tar -czf "coolify-cli-${VERSION}.tar.gz" coolify-*
cd ..

echo "✅ Release $VERSION created!"
echo "Archive: dist/coolify-cli-${VERSION}.tar.gz"
