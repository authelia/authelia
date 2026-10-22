# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.11.7-builder@sha256:8910fe662f448b860e8d16ef1c9a4f8946d3193bdc3e971bf0c55c9a2cf246f6 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.7@sha256:3095ad245c353e3bdb27a38f22a095c1cfa9e295ab21c0214c3e72b0e2960525

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
