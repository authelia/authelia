FROM caddy:2.11.2-builder@sha256:930cbe20ab83d080fc0f6f0de7ddc90840ee94f9dd1ec1725e70f938ffa715dc AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.2@sha256:25cdc846626b62d05f6b633b9b40c2c9f6ef89b515dc76133cefd920f7dbe562

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
