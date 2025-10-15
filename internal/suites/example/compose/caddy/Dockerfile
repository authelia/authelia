FROM caddy:2.10.2-builder@sha256:f8930c8c45d19fa60db1599500b39fcef775003ab1cace65737eb2cf9c06b1a4 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:786b7527345b82faf39e664ea26c71cc57bcb3a89bb2eada94bf0b2dc5304f0a

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
