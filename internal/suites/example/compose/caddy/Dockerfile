FROM caddy:2.11.4-builder@sha256:eb37f58646a901dc7727cf448cae36daaefaba79de33b5058dab79aa4c04aefb AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.4@sha256:eb37f58646a901dc7727cf448cae36daaefaba79de33b5058dab79aa4c04aefb

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
