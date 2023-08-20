#!/bin/sh

[[ ! -z ${UMASK} ]] && umask ${UMASK}

if [[ ! -z ${1} ]] && [[ ${1} != "--config" ]]; then
  exec "$@"
elif [[ $(id -u) != 0 ]] || [[ $(id -g) != 0 ]]; then
  exec authelia "$@"
else
  chown -R ${PUID}:${PGID} /config
  exec su-exec ${PUID}:${PGID} authelia "$@"
fi
