#!/bin/bash

DC_SCRIPT=./scripts/example/dc-example.sh

start_services() {
    $DC_SCRIPT up -d redis openldap authelia nginx nginx-tests
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
  echo "Prepare nginx-test configuration"
  cat example/nginx/nginx.conf | sed 's/listen 443 ssl/listen 8080 ssl/g' | dd of="test/integration/nginx.conf"
  
  echo "Build services images..."
  $DC_SCRIPT build
  
  echo "Start services..."
  start_services
  docker ps -a
  
  echo "Display services logs..."
  $DC_SCRIPT logs redis
  $DC_SCRIPT logs openldap
  $DC_SCRIPT logs nginx
  $DC_SCRIPT logs nginx-tests
  $DC_SCRIPT logs authelia
  
  echo "Check number of services"
  expect_services_count 5
  
  echo "Run integration tests..."
  $DC_SCRIPT run --rm integration-tests
  
  echo "Shutdown services..."
  shut_services
}

run_system_tests() {
  echo "Start services..."
  start_services
  expect_services_count 5
  
  ./node_modules/.bin/mocha --compilers ts:ts-node/register --recursive test/system
  shut_services  
}

run_other_tests() {
  echo "Test dev environment deployment (commands in README)"
  npm install --only=dev
  ./node_modules/.bin/grunt build-dist
  ./scripts/example/deploy-example.sh
  expect_services_count 4
}





set -e

echo "Make sure services are not already running"
shut_services

# Prepare & run integration tests
run_integration_tests

# Prepare & test example from end user perspective
run_system_tests

# Other tests like executing the deployment script
run_other_tests
