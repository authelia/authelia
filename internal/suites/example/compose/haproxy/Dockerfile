# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM haproxy:3.4.8-alpine@sha256:0f4ac017b3a028170f9727f0f0865cfae9264519efa5b7c7d31606ce4777f8bc
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
