FROM caddy:2.9.1-builder@sha256:1609bfce85bd4452a875e4d459f25e602c484b5a36e9c015511b5bdbd3539784 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:748016f285ed8c43a9ce6e3aed6d92d3009d90ca41157950880f40beaf3ff62b

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
