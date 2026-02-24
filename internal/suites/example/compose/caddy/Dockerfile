FROM caddy:2.11.1-builder@sha256:4f2f7dc18c7451f8b560dde511f76bf69d3db63350485f919616c8f4397497d0 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.1@sha256:9068f76202c0a03545036d32bf2d424d3b46c1174f254070f605002a2dbc9657

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
