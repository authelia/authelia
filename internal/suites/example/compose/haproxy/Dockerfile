# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM haproxy:3.4.7-alpine@sha256:09ba5c0b582cda2fa00e1f7f23f6e29cc27000089d6ffeb3c8683f930126ed10
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
