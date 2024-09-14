FROM caddy:2.8.4-builder@sha256:b1ee6157da9ce89796a6ce266713d701b383a17db2fa9cdb6dac04ef157544c8 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:c6a5f2835a53fd9e2b5e0623dc7ebf0f89c163c21c05cbbc35e4ea31ff424417

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
