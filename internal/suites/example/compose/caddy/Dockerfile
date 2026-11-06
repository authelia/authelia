# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.11.8-builder@sha256:d4eb070c36154bd2d4c94b40359a03c6168346d5af5404963c74691ab77fb63f AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.8@sha256:4fcafa98afd1c2e642e353e3c5cf9ddbe892fcee14bc990e7ba2c5b78e51f9cf

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
