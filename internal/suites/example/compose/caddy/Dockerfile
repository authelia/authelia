FROM caddy:2.10.2-builder@sha256:ff518a779d2f5df938d60a821f61929472ceef180ffb9d410da12a4290e1f04a AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:c1ccc35f8ba2567b90b72c88da31f68dc511ad6424e3885e911f2fd3fdc45ea3

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
