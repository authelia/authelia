#!/bin/sh

cmd=authelia
currentUser=$(whoami 2>&1)

if [[ "${currentUser}" == "root" ]]; then
  if [[ ! -z ${PUID:-} ]]; then
    chown -R ${PUID:-0} /config
  fi
  if [[ ! -z ${PGID:-} ]]; then
    chgrp -R ${PGID:-0} /config
  fi
  if [[ ! -z ${PUID:-} ]] || [[ ! -z ${PGID:-} ]]; then
    cmd="su-exec ${PUID:-0}:${PGID:-0} authelia"
  fi
fi

if [[ ! -z ${1:-} ]] && [[ "${1:0:1}" != "-" ]]; then
  exec "$@"
else
  exec ${cmd} "$@"
fi
