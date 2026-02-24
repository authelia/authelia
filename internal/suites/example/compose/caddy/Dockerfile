FROM caddy:2.11.1-builder@sha256:1c8dc3d45e6871a24ad3b023cf39ded7a5435159dce2ea3416773a75331ac92f AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.1@sha256:7f5a0f0468688c5305f1caf56302efb3d2512668b73d4d4eb2202ad53ae95b66

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
