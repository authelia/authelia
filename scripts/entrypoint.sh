#!/bin/sh

chown -R ${PUID:-0}:${PGID:-0} /logs /config

if [[ ! -z ${1:-} ]] && [[ "${1:0:1}" != "-" ]]; then
  exec "$@"
else
  exec su-exec ${PUID:-0}:${PGID:-0} authelia "$@"
fi
