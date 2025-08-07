FROM caddy:2.10.0-builder@sha256:d300ab11c67d279f272a6b8420bc381a66ac696a16c2a7aa55ea0262e705d78e AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:e23538fceb12f3f8cc97a174844aa99bdea7715023d6e088028850fd0601e2e2

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
