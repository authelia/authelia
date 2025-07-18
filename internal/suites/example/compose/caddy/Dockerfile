FROM caddy:2.10.0-builder@sha256:c1e0d8e7fe650596356c46319c8c47890045bf3cd4c3d2fc6463a7ff541a99c8 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:e23538fceb12f3f8cc97a174844aa99bdea7715023d6e088028850fd0601e2e2

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
