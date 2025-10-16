FROM caddy:2.10.2-builder@sha256:e8ec35d0787c123c0776d5e22570f2eeba76a730390104113554c271c40d2aeb AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:786b7527345b82faf39e664ea26c71cc57bcb3a89bb2eada94bf0b2dc5304f0a

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
