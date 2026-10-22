# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.11.7-builder@sha256:671d2a6e5465cc894c9ca57706c7ea7d101018059fe3154c65805a0eda550467 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.7@sha256:3095ad245c353e3bdb27a38f22a095c1cfa9e295ab21c0214c3e72b0e2960525

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
