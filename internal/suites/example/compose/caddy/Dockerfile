FROM caddy:2.10.0-builder@sha256:9edca605c07c8b5425d1985b4d4a1796329b11c3eba0b55f938e01916dcd96c8 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:30ccf0cb027e1d06cd6e453c04fc1c8eec665629b22ed69602c14c8a0512ead0

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
