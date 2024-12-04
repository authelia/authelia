FROM caddy:2.8.4-builder@sha256:5b1b97d37f69d58005b731b88b729a12f19e7b563291a1a2aa35cefff86f78d7 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:07417b42b03b8929ec80240d1a83e27a9c1346ddbff65988608fd1757269c1eb

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
