# SPDX-FileCopyrightText: 2019 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.6.4-builder AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.6.4

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
