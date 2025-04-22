FROM caddy:2.10.0-builder@sha256:680deea9ca2c0e5e415573a0debd1aa92a7a298a147e70079fd5ac6eaaadbad9 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:e759110e56bae353dbceddff9d7665feb5229d5afac1a5e7e3f42d99218f9ba6

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
