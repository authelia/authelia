#!/bin/sh

source /app/.healthcheck.env

if [ -z "${X_AUTHELIA_HEALTHCHECK}" ]; then
  exit 0
fi

if [ -z "${X_AUTHELIA_HEALTHCHECK_SCHEME}" ]; then
  X_AUTHELIA_HEALTHCHECK_SCHEME=http
fi

if [ -z "${X_AUTHELIA_HEALTHCHECK_HOST}" ]; then
  X_AUTHELIA_HEALTHCHECK_HOST=localhost
fi

if [ -z "${X_AUTHELIA_HEALTHCHECK_PORT}" ]; then
  X_AUTHELIA_HEALTHCHECK_PORT=9091
fi

wget --quiet --no-check-certificate --tries=1 --spider "${X_AUTHELIA_HEALTHCHECK_SCHEME}://${X_AUTHELIA_HEALTHCHECK_HOST}:${X_AUTHELIA_HEALTHCHECK_PORT}${X_AUTHELIA_HEALTHCHECK_PATH}/api/health" || exit 1
