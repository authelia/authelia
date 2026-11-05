# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.11.8-builder@sha256:eb37f58646a901dc7727cf448cae36daaefaba79de33b5058dab79aa4c04aefb AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.8@sha256:f04f1e8d78da0c65ec6f81b4ea8d216849775f3bc3c87790d95fcb2ec3722f12

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
