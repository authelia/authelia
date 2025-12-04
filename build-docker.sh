#!/usr/bin/env bash

set -e

# Configuration
IMAGE_NAME="ghcr.io/deep-jiwan/mypackages/authelia/customtheme"
IMAGE_TAG="autobuilddocker"
FULL_IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Building Authelia Custom Theme Docker${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${YELLOW}Error: Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

echo -e "${GREEN}Building Docker image...${NC}"
echo "This will build both frontend and backend inside Docker."
echo "Note: Go build may take 10-15 minutes and use significant memory."
echo ""

docker build \
    --file Dockerfile.custom \
    --tag "${FULL_IMAGE}" \
    --memory="2g" \
    --memory-swap="4g" \
    --progress=plain \
    .

echo ""
echo -e "${GREEN}Tagging image...${NC}"
docker tag "${FULL_IMAGE}" "${IMAGE_NAME}:latest"

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✓ Build completed successfully!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Image tags created:"
echo "  - ${FULL_IMAGE}"
echo "  - ${IMAGE_NAME}:latest"
echo ""
echo "To run the container:"
echo "  docker run -d \\"
echo "    --name authelia \\"
echo "    -p 9091:9091 \\"
echo "    -v ./config.yml:/config/configuration.yml:ro \\"
echo "    ${FULL_IMAGE}"
echo ""
echo "To push to GitHub Container Registry:"
echo "  docker push ${FULL_IMAGE}"
echo "  docker push ${IMAGE_NAME}:latest"
echo ""
