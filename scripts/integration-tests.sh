#!/bin/bash

set -e

echo "Make sure services are not already running"
./scripts/dc-example.sh down

echo "Prepare nginx-test configuration"
cat example/nginx/nginx.conf | sed 's/listen 443 ssl/listen 8080 ssl/g' | dd of="test/integration/nginx.conf"

echo "Build services images..."
./scripts/dc-example.sh build

echo "Start services..."
./scripts/dc-example.sh up -d redis openldap 
sleep 2
./scripts/dc-example.sh up -d authelia nginx nginx-tests
sleep 3
docker ps -a

echo "Display services logs..."
./scripts/dc-example.sh logs redis
./scripts/dc-example.sh logs openldap
./scripts/dc-example.sh logs nginx
./scripts/dc-example.sh logs nginx-tests
./scripts/dc-example.sh logs authelia

echo "Check number of services"
./scripts/check-services.sh

echo "Run integration tests..."
./scripts/dc-example.sh run --rm integration-tests

echo "Shutdown services..."
./scripts/dc-example.sh down
