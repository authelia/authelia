FROM caddy:2.11.2-builder@sha256:84dfc3479309c690643ada9279b3e0b4352ce56b0ec8fd802c668f42b546e98f AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.2@sha256:1e40b251ca9639ead7b5cd2cedcc8765adfbabb99450fe23f130eefabf50f4bc

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
