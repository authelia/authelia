FROM caddy:2.10.2-builder@sha256:ff518a779d2f5df938d60a821f61929472ceef180ffb9d410da12a4290e1f04a AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:86d2ab93167b0c24f9f86e8a4214dee0ba25f86808002bfd29a78677824a2483

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
