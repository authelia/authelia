#!/bin/sh

# SPDX-FileCopyrightText: 2019 Authelia
#
# SPDX-License-Identifier: Apache-2.0

if [[ ! -z ${1} ]] && [[ ${1} != "--config" ]]; then
  exec "$@"
elif [[ $(id -u) != 0 ]] || [[ $(id -g) != 0 ]]; then
  exec authelia "$@"
else
  chown -R ${PUID}:${PGID} /config
  exec su-exec ${PUID}:${PGID} authelia "$@"
fi