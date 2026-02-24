FROM caddy:2.11.1-builder@sha256:9780b126deae015b2c262fa9f4f1f951577e1f6a5fe4d88c0573aeb622071b2d AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.1@sha256:01511babb51d90619d4a9919206d45b3e7fef645ba1ca9ff0fc298b4c80fac16

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
