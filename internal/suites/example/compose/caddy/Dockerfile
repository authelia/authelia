FROM caddy:2.10.0-builder@sha256:e5612b84a4a8c0d452da55d3d7d97985a5039ef3be6e399fd7264bb070f30559 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:c10cf853496a180595e521a7e5081e0ba0a85e4bd54321a751e3dc91c7b3dac9

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
