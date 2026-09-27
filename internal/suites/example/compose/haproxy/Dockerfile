# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM haproxy:3.4.5-alpine@sha256:7ffdd3845020aa4c97ffaa56c8cab3ad473ecae0b8ea5e55d2378d663755e092
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
