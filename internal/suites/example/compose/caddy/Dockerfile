FROM caddy:2.7.5-builder AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.7.5

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
