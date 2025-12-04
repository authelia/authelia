#!/bin/bash

# run authelia with build front end and backend.

# Authelia Development Runner
# Usage: ./run-dev.sh [web|go|web go|<empty>]
#   web go  - Build both frontend and backend, then run
#   web     - Build frontend only
#   go      - Build backend only
#   <empty> - Run without building (uses existing binaries)

set -e

# Use system Node.js 22
export PATH=/usr/bin:$PATH

BUILD_WEB=false
BUILD_GO=false
RUN_AUTHELIA=false

# Parse arguments
if [ $# -eq 0 ]; then
    # No arguments: just run
    RUN_AUTHELIA=true
elif [ "$1" == "web" ] && [ "$2" == "go" ]; then
    # Build both
    BUILD_WEB=true
    BUILD_GO=true
    RUN_AUTHELIA=true
elif [ "$1" == "go" ] && [ "$2" == "web" ]; then
    # Build both (order doesn't matter)
    BUILD_WEB=true
    BUILD_GO=true
    RUN_AUTHELIA=true
elif [ "$1" == "web" ]; then
    # Build web only
    BUILD_WEB=true
elif [ "$1" == "go" ]; then
    # Build go only
    BUILD_GO=true
else
    echo "Usage: $0 [web|go|web go|<empty>]"
    echo ""
    echo "Examples:"
    echo "  $0           - Run Authelia without building"
    echo "  $0 web go    - Build both frontend and backend, then run"
    echo "  $0 web       - Build frontend only"
    echo "  $0 go        - Build backend only"
    exit 1
fi

# Build Web (Frontend)
if [ "$BUILD_WEB" = true ]; then
    echo "=========================================="
    echo "Building Frontend (React + Vite)"
    echo "=========================================="
    echo "Node.js version: $(node --version)"
    echo ""
    
    # Check if pnpm is installed
    if ! command -v pnpm &> /dev/null; then
        echo "Installing pnpm..."
        npm install -g pnpm
    fi
    
    cd web
    echo "→ Installing dependencies..."
    pnpm install --frozen-lockfile
    echo "→ Building frontend..."
    pnpm build
    cd ..
    
    echo "→ Copying API documentation..."
    cp -r api internal/server/public_html/
    
    echo "✓ Frontend build complete"
    echo ""
fi

# Build Go (Backend)
if [ "$BUILD_GO" = true ]; then
    echo "=========================================="
    echo "Building Backend (Go)"
    echo "=========================================="
    echo "Go version: $(go version)"
    echo ""
    
    echo "→ Building backend..."
    go build -tags dev -o authelia ./cmd/authelia
    
    echo "✓ Backend build complete"
    echo ""
fi

# Run Authelia
if [ "$RUN_AUTHELIA" = true ]; then
    if [ ! -f "./authelia" ]; then
        echo "Error: authelia binary not found. Build it first with: $0 go"
        exit 1
    fi
    
    echo "=========================================="
    echo "Starting Authelia"
    echo "=========================================="
    echo "Port: 9010"
    echo "Config: config.template.yml"
    echo "Users: admin:password, user:password"
    echo "Web UI: http://localhost:9010"
    echo ""
    
    ./authelia --config config.template.yml
fi





# cd /home/adgpi0/authelia/authelia/web
# export PATH=/usr/bin:$PATH
# pnpm start



 