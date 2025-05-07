FROM caddy:2.10.0-builder@sha256:3df563101e58f6d1c6f07777489ba427c2ee6495f05c6e580612c70200346c25 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:e759110e56bae353dbceddff9d7665feb5229d5afac1a5e7e3f42d99218f9ba6

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
