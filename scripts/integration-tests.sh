#!/bin/bash

DC_SCRIPT=./scripts/example/dc-example.sh

run_services() {
    $DC_SCRIPT up -d redis openldap 
    sleep 2
    $DC_SCRIPT up -d authelia nginx nginx-tests
    sleep 3
}

set -e

echo "Make sure services are not already running"
$DC_SCRIPT down


# Prepare & run integration tests

echo "Prepare nginx-test configuration"
cat example/nginx/nginx.conf | sed 's/listen 443 ssl/listen 8080 ssl/g' | dd of="test/integration/nginx.conf"

echo "Build services images..."
$DC_SCRIPT build

echo "Start services..."
run_services
docker ps -a

echo "Display services logs..."
$DC_SCRIPT logs redis
$DC_SCRIPT logs openldap
$DC_SCRIPT logs nginx
$DC_SCRIPT logs nginx-tests
$DC_SCRIPT logs authelia

echo "Check number of services"
./scripts/example/check-services.sh

echo "Run integration tests..."
$DC_SCRIPT run --rm integration-tests

echo "Shutdown services..."
$DC_SCRIPT down

# Prepare & test example from end user perspective

echo "Start services..."
run_services

./node_modules/.bin/mocha --compilers ts:ts-node/register --recursive test/system

$DC_SCRIPT down

