FROM caddy:2.5.0-builder AS builder

RUN xcaddy build proxy-pre-check

FROM caddy:2.5.0

COPY --from=builder /usr/bin/caddy /usr/bin/caddy