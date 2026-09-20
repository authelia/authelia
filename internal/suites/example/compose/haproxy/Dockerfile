# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM haproxy:3.4.4-alpine@sha256:52c5921e1619f39cbd5b25e1b4b5847667917f39745056cf004d9c263fbf11b9
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
