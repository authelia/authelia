# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM haproxy:3.4.7-alpine@sha256:3c8c1df0ae3bea8834b3890fa6fa3b1653b4d6d9c44aecefba46428bfcfde2c0
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
