#!/bin/sh

source /app/.healthcheck.env

if [ -z "${HEATHCHECK_SCHEME}" ]; then
  HEATHCHECK_SCHEME=http
fi

if [ -z "${HEATHCHECK_HOST}" ] || [ "${HEATHCHECK_HOST}" = "0.0.0.0" ]; then
  HEATHCHECK_HOST=localhost
fi

if [ -z "${HEATHCHECK_PORT}" ]; then
  HEATHCHECK_PORT=9091
fi

wget --quiet --no-check-certificate --tries=1 --spider "${HEATHCHECK_SCHEME}://${HEATHCHECK_HOST}:${HEATHCHECK_PORT}${HEATHCHECK_PATH}/api/health" || exit 1
