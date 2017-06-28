#!/bin/bash

set -e

echo "Build services images..."
./scripts/dc-test.sh build

echo "Start services..."
./scripts/dc-test.sh up -d authelia nginx openldap
sleep 3
docker ps -a

echo "Display services logs..."
./scripts/dc-test.sh logs authelia
./scripts/dc-test.sh logs nginx
./scripts/dc-test.sh logs openldap

echo "Run integration tests..."
./scripts/dc-test.sh run --rm --name int-test int-test

echo "Shutdown services..."
./scripts/dc-test.sh down
