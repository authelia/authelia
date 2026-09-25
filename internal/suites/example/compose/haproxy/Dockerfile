# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM haproxy:3.4.5-alpine@sha256:4ebeea8b11f17f1fe044fba65cbc42f94e478d6d3ee223f59cacaf2621ecf8a1
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
