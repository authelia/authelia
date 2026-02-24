FROM caddy:2.11.1-builder@sha256:4f2f7dc18c7451f8b560dde511f76bf69d3db63350485f919616c8f4397497d0 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.1@sha256:7f5a0f0468688c5305f1caf56302efb3d2512668b73d4d4eb2202ad53ae95b66

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
