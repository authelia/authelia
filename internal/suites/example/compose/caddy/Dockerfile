FROM caddy:2.10.0-builder@sha256:c9fd63066cb907c05488886402ed9869ba8b90741a6af2d85b33ab43ba94f342 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:e759110e56bae353dbceddff9d7665feb5229d5afac1a5e7e3f42d99218f9ba6

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
