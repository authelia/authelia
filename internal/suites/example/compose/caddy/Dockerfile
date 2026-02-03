FROM caddy:2.10.2-builder@sha256:23e70f63a88bacd9b0564d055eed5c25b98e90930876cca01aee1f5d1ae29748 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:f20f80e1fb627294fb84b8515b7593aff8018c840f1396dc942a50ed0c2db648

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
