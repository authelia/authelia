FROM caddy:2.11.4-builder@sha256:26f5271c26850df54ef7a9f3704b5d688c1ad51b5cde509c9d011b4e95cc8397 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.4@sha256:af5fdcd76f2db5e4e974ee92f96ee8c0fc3edb55bd4ba5032547cbf3f65e486d

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
