FROM caddy:2.10.0-builder@sha256:6e7a8ab47f8663a71e07458bf7f58e258fda81697a5af99e9eb836d9341a953a AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:c5876b163e84c44815e2fbba68245367dcf341a15947f80bffffa011bdc90ece

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
