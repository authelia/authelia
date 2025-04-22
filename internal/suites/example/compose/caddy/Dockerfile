FROM caddy:2.10.0-builder@sha256:5b9518da6d79ae9960f09ba1175a2bc2e90614ad9c062df26fe8bdf5a95d2f59 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:0bb77a80ca42cc08c193a8311de01bd06aee4a5d5fd8d102d320982aa2ed17c0

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
