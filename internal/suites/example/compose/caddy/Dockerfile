# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.11.9-builder@sha256:1b828679ae08622c534d8e83ed80c6c25d343202f0586fe3b52d8e0ce33413b9 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.9@sha256:e40b94d0a43b5ff69cb20568d73f10c70703df7b0ffc2298eb3d723cdc9d0bd1

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
