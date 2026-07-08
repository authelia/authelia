FROM caddy:2.11.4-builder@sha256:feb77c5a25fd44d448809bb4f108368a66f1899e61acb2f7f023d290423fe3e9 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.4@sha256:af5fdcd76f2db5e4e974ee92f96ee8c0fc3edb55bd4ba5032547cbf3f65e486d

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
