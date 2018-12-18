#!/bin/bash

DC_SCRIPT=./scripts/example-commit/dc-example.sh
EXPECTED_SERVICES_COUNT=8

build_services() {
    $DC_SCRIPT build authelia
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
  ./node_modules/.bin/grunt test-int
}

run_other_tests() {
  echo "Test dev environment deployment (commands in README)"
  ./scripts/example-commit/deploy-example.sh
  expect_services_count $EXPECTED_SERVICES_COUNT
  ./scripts/example-commit/undeploy-example.sh
}

run_other_tests_docker() {
  echo "Test dev docker deployment (commands in README)"
  ./scripts/example-dockerhub/deploy-example.sh
  expect_services_count $EXPECTED_SERVICES_COUNT
  ./scripts/example-dockerhub/undeploy-example.sh
}

set -e

# Build the container
build_services

# Pull all images
$DC_SCRIPT pull

# Prepare & test example from end user perspective
run_integration_tests

# Other tests like executing the deployment script
run_other_tests

# Test example with precompiled container
run_other_tests_docker
