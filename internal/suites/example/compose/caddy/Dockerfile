# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.11.7-builder@sha256:91443bd37eafc9ea2c3ff43749e33e56a61a323da1a9cf36f6d25f0f5c468027 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.7@sha256:f2a1290d0463aad60660d4ec134943f183ee2a5f6c3eb7bf32dd984f2f020772

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
