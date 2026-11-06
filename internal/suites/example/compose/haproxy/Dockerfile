# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM haproxy:3.4.7-alpine@sha256:5312d37032c079c4937d45ef63758852434351f3b425a3703521fd999837abcf
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
