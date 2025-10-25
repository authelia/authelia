FROM caddy:2.10.2-builder@sha256:53f91ad7c5f1ab9a607953199b7c1e10920c570ae002aef913d68ed7464fb19f AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:786b7527345b82faf39e664ea26c71cc57bcb3a89bb2eada94bf0b2dc5304f0a

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
