#!/bin/bash

DC_SCRIPT=./scripts/example-commit/dc-example.sh
EXPECTED_SERVICES_COUNT=5

start_services() {
    $DC_SCRIPT up -d mongo redis openldap authelia nginx
    sleep 3
}

shut_services() {
  $DC_SCRIPT down
}

expect_services_count() {
  EXPECTED_COUNT=$1
  service_count=`docker ps -a | grep "Up " | wc -l`
  
  if [ "${service_count}" -eq "$EXPECTED_COUNT" ]
  then
    echo "Services are up and running."
  else
    echo "Some services exited..."
    docker ps -a
    exit 1
  fi
}

run_integration_tests() {
  echo "Start services..."
  start_services
  expect_services_count $EXPECTED_SERVICES_COUNT 
  
  sleep 5
  ./node_modules/.bin/grunt run:integration-tests
  shut_services  
}

run_other_tests() {
  echo "Test dev environment deployment (commands in README)"
  npm install --only=dev
  ./node_modules/.bin/grunt build-dist
  ./scripts/example-commit/deploy-example.sh
  expect_services_count 5
}

run_other_tests_docker() {
  echo "Test dev docker deployment (commands in README)"
  ./scripts/example-dockerhub/deploy-example.sh
  expect_services_count 5
}





set -e

echo "Make sure services are not already running"
shut_services

# Prepare & test example from end user perspective
run_integration_tests

# Other tests like executing the deployment script
run_other_tests

# Test example with precompiled container
run_other_tests_docker