FROM caddy:2.10.2-builder@sha256:472b384553e37f446d82d73c70340567da67fd6fb7386d4514dcb5d8c670c90d AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:f20f80e1fb627294fb84b8515b7593aff8018c840f1396dc942a50ed0c2db648

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
