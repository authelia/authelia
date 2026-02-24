FROM caddy:2.11.1-builder@sha256:1c8dc3d45e6871a24ad3b023cf39ded7a5435159dce2ea3416773a75331ac92f AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.1@sha256:01511babb51d90619d4a9919206d45b3e7fef645ba1ca9ff0fc298b4c80fac16

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
