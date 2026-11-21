# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM haproxy:3.4.8-alpine@sha256:8aa6458a9d22a8364a2d2ee9ce93de0cc4a2789e3cf63f0b44adc972863724d7
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
