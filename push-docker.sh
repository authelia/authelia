#!/usr/bin/env bash


# Script to push Docker images to GitHub Container Registry

set -e

# Configuration
IMAGE_NAME="ghcr.io/deep-jiwan/mypackages/authelia/customtheme"
IMAGE_TAG="autobuild"
FULL_IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Push to GitHub Container Registry${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if logged in to GitHub Container Registry
echo "Checking GitHub Container Registry authentication..."
if ! docker info 2>/dev/null | grep -q "Username"; then
    echo -e "${YELLOW}You need to login to GitHub Container Registry first:${NC}"
    echo ""
    echo "  1. Create a Personal Access Token at:"
    echo "     https://github.com/settings/tokens"
    echo "     (With 'write:packages' permission)"
    echo ""
    echo "  2. Login with:"
    echo "     echo YOUR_TOKEN | docker login ghcr.io -u Deep-Jiwan --password-stdin"
    echo ""
    read -p "Press Enter after you've logged in, or Ctrl+C to cancel..."
fi

# Check if image exists
if ! docker image inspect "${FULL_IMAGE}" > /dev/null 2>&1; then
    echo -e "${YELLOW}Error: Image ${FULL_IMAGE} not found.${NC}"
    echo "Please run ./build-docker.sh first."
    exit 1
fi

echo ""
echo -e "${GREEN}Pushing ${FULL_IMAGE}...${NC}"
docker push "${FULL_IMAGE}"

echo ""
echo -e "${GREEN}Pushing ${IMAGE_NAME}:latest...${NC}"
docker push "${IMAGE_NAME}:latest"

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✓ Push completed successfully!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Your images are now available at:"
echo "  - https://github.com/Deep-Jiwan/mypackages/pkgs/container/authelia%2Fcustomtheme"
echo ""
