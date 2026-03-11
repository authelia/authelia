FROM caddy:2.11.2-builder@sha256:84dfc3479309c690643ada9279b3e0b4352ce56b0ec8fd802c668f42b546e98f AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.2@sha256:261437d35b88cc2f8fa1ed337ebb6b6ce2bd5cd61b2a28cdf83cb201db0a853b

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
