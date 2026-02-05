#!/bin/bash

# Configuration
IMAGE_NAME="acecandy/fasaxi-linker"
VERSION="latest"

# Build Image
echo "🚀 Building Docker image: $IMAGE_NAME:$VERSION"
docker buildx build --platform linux/arm64,linux/amd64 -t $IMAGE_NAME:$VERSION -f Dockerfile  --push .

echo ""
echo "✅ Build complete!"
echo ""
echo "✅ To push to Docker Hub! $IMAGE_NAME:$VERSION"
