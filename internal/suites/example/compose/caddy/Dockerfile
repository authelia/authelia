# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.11.4-builder@sha256:401121e61853cb9c7df83cba43e532e49a68e30f32921c7bdcb7a4912168c067 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.4@sha256:0c994536bddb66445885237f1a5dcc1916bccea922661c76b4e9fc24061f9b52

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
