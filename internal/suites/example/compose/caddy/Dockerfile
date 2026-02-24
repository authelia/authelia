FROM caddy:2.11.1-builder@sha256:4f2f7dc18c7451f8b560dde511f76bf69d3db63350485f919616c8f4397497d0 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.1@sha256:01511babb51d90619d4a9919206d45b3e7fef645ba1ca9ff0fc298b4c80fac16

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
