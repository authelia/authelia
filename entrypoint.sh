#!/bin/sh

[ -n "${UMASK}" ] && umask "${UMASK}"

if [ -n "${1}" ] && [ "${1}" != "--config" ]; then
  exec "${@}"
elif [ "$(id -u)" != 0 ] || [ "$(id -g)" != 0 ]; then
  exec authelia "${@}"
else
  chown -R "${PUID}:${PGID}" /config
  exec su-exec "${PUID}:${PGID}" authelia "${@}"
fi
