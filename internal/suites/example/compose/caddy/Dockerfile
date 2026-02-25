FROM caddy:2.11.1-builder@sha256:9780b126deae015b2c262fa9f4f1f951577e1f6a5fe4d88c0573aeb622071b2d AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.1@sha256:9068f76202c0a03545036d32bf2d424d3b46c1174f254070f605002a2dbc9657

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
