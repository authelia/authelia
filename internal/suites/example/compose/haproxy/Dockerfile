# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM haproxy:3.4.6-alpine@sha256:a235ee03fe88d9fd3e1a91ee568ff48d581f6866e6148005840fb8c362d358ab
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
