FROM caddy:2.10.0-builder@sha256:6e7a8ab47f8663a71e07458bf7f58e258fda81697a5af99e9eb836d9341a953a AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:f880657326733fb266a317b4ee31d6a3090a6bbff3dcb0c97833093810d59b46

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
