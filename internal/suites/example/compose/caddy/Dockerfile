FROM caddy:2.10.0-builder@sha256:acf9b51367f2cdd94a5621b1d5f37a3d095b5f6e2157c83b48d2b1f127576366 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:e23538fceb12f3f8cc97a174844aa99bdea7715023d6e088028850fd0601e2e2

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
