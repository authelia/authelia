# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM haproxy:3.4.7-alpine@sha256:dc532308825d207e294624eda1c87de5b6dbd0e6f1f154a3444c977ebb0b9c84
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
