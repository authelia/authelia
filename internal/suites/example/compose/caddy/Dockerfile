FROM caddy:2.7.6-builder AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.7.6

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
