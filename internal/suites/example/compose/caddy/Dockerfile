FROM caddy:2.11.4-builder@sha256:2cffe63dc72e3ea610f70a6bd4b11a4d1c873000e191f1bf59aae88d464906c5 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.4@sha256:1557580dffd3f309e30bd76b23a26aa3970d98bff11313cfa3a63b735ac8d04a

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
