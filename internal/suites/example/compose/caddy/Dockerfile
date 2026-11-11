# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.11.9-builder@sha256:1b828679ae08622c534d8e83ed80c6c25d343202f0586fe3b52d8e0ce33413b9 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.9@sha256:0c2243efcf437ab7f7a978f1027fcf0d4f9be8a75df27b1b3fbc5dbe0ffa8cf0

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
