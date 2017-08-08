#!/bin/bash

set -e

docker-compose \
  -f docker-compose.base.yml \
  -f example/authelia/docker-compose.yml \
  -f example/mongo/docker-compose.yml \
  -f example/redis/docker-compose.yml \
  -f example/nginx/docker-compose.yml \
  -f example/ldap/docker-compose.yml $*
