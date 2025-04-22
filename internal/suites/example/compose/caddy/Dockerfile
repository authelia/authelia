FROM caddy:2.10.0-builder@sha256:5b9518da6d79ae9960f09ba1175a2bc2e90614ad9c062df26fe8bdf5a95d2f59 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:e759110e56bae353dbceddff9d7665feb5229d5afac1a5e7e3f42d99218f9ba6

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
