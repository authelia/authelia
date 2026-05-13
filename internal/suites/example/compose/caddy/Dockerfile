FROM caddy:2.11.3-builder@sha256:090dbc9541137fc6caac753baf94491d9998d7a04a3facb057dbfe8ddef8739a AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.3@sha256:a464e5c4fb9c324b53adca428a40a66631e59127153b8816fab58103817ba8f0

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
