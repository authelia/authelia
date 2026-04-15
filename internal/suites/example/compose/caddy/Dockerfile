FROM caddy:2.11.2-builder@sha256:ccd4503985eaa4aac67dcc75632ab9e9df5bfb55241c95dcd7654f7e986883d4 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.2@sha256:1e40b251ca9639ead7b5cd2cedcc8765adfbabb99450fe23f130eefabf50f4bc

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
