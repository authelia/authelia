#!/bin/bash

set -e

docker --version
docker-compose --version

# Run unit tests
grunt test-unit

# Build the app from Typescript and package
grunt build-dist

# Run integration/example tests
./scripts/integration-tests.sh

# Test npm deployment before actual deployment
./scripts/npm-deployment-test.sh
