FROM caddy:2.10.0-builder@sha256:6e7a8ab47f8663a71e07458bf7f58e258fda81697a5af99e9eb836d9341a953a AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:cd038960363e661b805a40ee6d03ba98608538f197c08cb915cb4af89e23a5b4

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
