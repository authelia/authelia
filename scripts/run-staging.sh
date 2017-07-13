#!/bin/bash

set -e

# Build production environment and set it up
./scripts/dc-example.sh build
./scripts/dc-example.sh up -d
 
# Wait for services to be running
sleep 5

# Check if services are correctly running
./scripts/check-services.sh

./scripts/dc-example.sh down
