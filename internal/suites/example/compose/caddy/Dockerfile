# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.11.7-builder@sha256:1ab914bd604996ab195f6326665cd2ae5d3df62feba973e776c4c8d5967d57a5 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.7@sha256:d41fbe1c07bfb1bed6336eff319c602a5fc7f69f3a92f84c231385a8730813ef

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
