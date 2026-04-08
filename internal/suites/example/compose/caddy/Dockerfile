FROM caddy:2.11.2-builder@sha256:a13471688cd43f00792536ecf577a1c1453dd9c8b2bfe236bf17b03c1ac5609d AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.2@sha256:1e40b251ca9639ead7b5cd2cedcc8765adfbabb99450fe23f130eefabf50f4bc

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
