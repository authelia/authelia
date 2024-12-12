FROM caddy:2.8.4-builder@sha256:dd54498b9546c06cdafd0e3b20276e6999f07102cc1fb578f0408c3bf8399393 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:07417b42b03b8929ec80240d1a83e27a9c1346ddbff65988608fd1757269c1eb

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
