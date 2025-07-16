FROM caddy:2.10.0-builder@sha256:c1e0d8e7fe650596356c46319c8c47890045bf3cd4c3d2fc6463a7ff541a99c8 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:c10cf853496a180595e521a7e5081e0ba0a85e4bd54321a751e3dc91c7b3dac9

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
