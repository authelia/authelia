#!/bin/bash

set -e

export PATH=./scripts:$PATH

docker --version
docker-compose --version
echo "node `node -v`"
echo "npm `npm -v`"

# Run unit tests
authelia-scripts test

# Build
authelia-scripts build

# Run integration/example tests
./scripts/integration-tests.sh

# Test npm deployment before actual deployment
# ./scripts/npm-deployment-test.sh
