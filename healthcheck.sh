#!/bin/sh

AUTHELIA_CONFIG=$(ps | grep authelia | awk '{print $6}' | head -1)
AUTHELIA_SCHEME=$(cat "${AUTHELIA_CONFIG}" | grep ^tls)
AUTHELIA_PORT=$(cat "${AUTHELIA_CONFIG}" | grep ^port | sed -e 's/port: //')

if [[ -z ${AUTHELIA_PORT} ]]; then
  AUTHELIA_PORT=9091
fi

if [[ -z ${AUTHELIA_SCHEME} ]]; then
  AUTHELIA_SCHEME=http
else
  AUTHELIA_SCHEME=https
fi

wget --quiet --tries=1 --spider ${AUTHELIA_SCHEME}://localhost:${AUTHELIA_PORT}/api/state || exit 1