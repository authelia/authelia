FROM caddy:2.10.2-builder@sha256:f8930c8c45d19fa60db1599500b39fcef775003ab1cace65737eb2cf9c06b1a4 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:61a257991a7cca3d843218fb4a4eaf8b3241443fed67f43b5a2b56848237eba0

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
