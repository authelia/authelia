FROM caddy:2.9.1-builder@sha256:1609bfce85bd4452a875e4d459f25e602c484b5a36e9c015511b5bdbd3539784 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:1972548f05660a069cfdf052aa2b923ce691627cbad1b7b97fb0a18533a766f2

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
