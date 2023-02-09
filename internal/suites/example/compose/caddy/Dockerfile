FROM caddy:2.6.3-builder AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.6.3

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
