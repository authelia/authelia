FROM caddy:2.10.2-builder@sha256:01668408cc26e2e00c9d067c30cb43b2ba14ad1f2808beda55503cb2a31f59dc AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:f20f80e1fb627294fb84b8515b7593aff8018c840f1396dc942a50ed0c2db648

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
