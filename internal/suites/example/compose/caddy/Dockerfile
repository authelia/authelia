FROM caddy:2.11.4-builder@sha256:ffa8572dfe784332cce850b255926465b5a71c947af60e6b4b17bf6adca4fedb AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.4@sha256:844f60b64e4724a5aa8245e019dace0d3f199f7433ce6c57676cb30a920dbad9

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
