FROM caddy:2.10.2-builder@sha256:b6424b4a90e25fde5cb9fd8e1da716159a313869ac3ba1c34b11c50781acab81 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:c1ccc35f8ba2567b90b72c88da31f68dc511ad6424e3885e911f2fd3fdc45ea3

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
