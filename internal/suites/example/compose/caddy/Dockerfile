FROM caddy:2.10.2-builder@sha256:00dfa5f39333243282c82d73de2acfe0609cb4a7879eb96946dcdfbf3a27ff87 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:f20f80e1fb627294fb84b8515b7593aff8018c840f1396dc942a50ed0c2db648

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
