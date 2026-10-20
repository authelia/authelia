# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM haproxy:3.4.7-alpine@sha256:251da69821897ec8c4819f12b1babbc4b71f562943addc9b9eb14bb6ece40dc6
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
