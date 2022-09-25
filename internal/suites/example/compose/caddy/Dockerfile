FROM caddy:2.6.1-builder AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.6.1

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
