FROM caddy:2.8.4-builder@sha256:5d96a90cf944be4b54659eb8d22da07296b199b82db9a3451dac17e182d24441 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:c6a5f2835a53fd9e2b5e0623dc7ebf0f89c163c21c05cbbc35e4ea31ff424417

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
