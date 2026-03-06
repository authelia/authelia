FROM caddy:2.11.1-builder@sha256:232766b340d6203e9f495ca9b4c6174cc1f75bde6ddeb47b7dc7979a2c5cda4e AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.1@sha256:9068f76202c0a03545036d32bf2d424d3b46c1174f254070f605002a2dbc9657

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
