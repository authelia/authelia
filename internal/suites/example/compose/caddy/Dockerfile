# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.11.4-builder@sha256:369218c81ca6d6af249981221b3a5c764d886dd5b058f51d144066de13f2418d AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.4@sha256:0c994536bddb66445885237f1a5dcc1916bccea922661c76b4e9fc24061f9b52

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
