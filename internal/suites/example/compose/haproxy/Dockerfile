# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM haproxy:3.4.7-alpine@sha256:2caaf8b1a1ef67f3f1ea9b721f2821675d3ba4500fb518e23884fecb93e8dbaa
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
