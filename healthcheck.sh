#!/bin/sh

source /app/.healthcheck.env

if [ -z "${HEALTHCHECK_SCHEME}" ]; then
  HEALTHCHECK_SCHEME=http
fi

if [ -z "${HEALTHCHECK_HOST}" ] || [ "${HEALTHCHECK_HOST}" = "0.0.0.0" ]; then
  HEALTHCHECK_HOST=localhost
fi

if [ -z "${HEALTHCHECK_PORT}" ]; then
  HEALTHCHECK_PORT=9091
fi

wget --quiet --no-check-certificate --tries=1 --spider "${HEALTHCHECK_SCHEME}://${HEALTHCHECK_HOST}:${HEALTHCHECK_PORT}${HEALTHCHECK_PATH}/api/health" || exit 1
