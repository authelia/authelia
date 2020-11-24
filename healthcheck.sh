#!/bin/sh

AUTHELIA_CONFIG=$(pgrep -af authelia | awk '{print $4}')
AUTHELIA_SCHEME=$(grep ^tls "${AUTHELIA_CONFIG}")
AUTHELIA_HOST=$(grep ^host "${AUTHELIA_CONFIG}" | sed -e 's/host: //' -e 's/\r//')
AUTHELIA_PORT=$(grep ^port "${AUTHELIA_CONFIG}" | sed -e 's/port: //' -e 's/\r//')

if [ -z "${AUTHELIA_SCHEME}" ]; then
  AUTHELIA_SCHEME=http
else
  AUTHELIA_SCHEME=https
fi

if [ -z "${AUTHELIA_HOST}" ] || [ "${AUTHELIA_HOST}" = "0.0.0.0" ]; then
  AUTHELIA_HOST=localhost
fi

if [ -z "${AUTHELIA_PORT}" ]; then
  AUTHELIA_PORT=9091
fi

wget --quiet --tries=1 --spider "${AUTHELIA_SCHEME}://${AUTHELIA_HOST}:${AUTHELIA_PORT}/api/state" || exit 1
