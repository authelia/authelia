FROM caddy:2.8.4-builder@sha256:6e9733694e9d257c215925b2429c6e9949114ef7a4205d9e4718aefbf15daa77 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:07417b42b03b8929ec80240d1a83e27a9c1346ddbff65988608fd1757269c1eb

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
