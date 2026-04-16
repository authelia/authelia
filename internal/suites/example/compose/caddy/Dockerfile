FROM caddy:2.11.2-builder@sha256:930cbe20ab83d080fc0f6f0de7ddc90840ee94f9dd1ec1725e70f938ffa715dc AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.2@sha256:2acb10cebb92eea91a40b76691aff73adde9151416facbeab630bbc66d0969ab

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
