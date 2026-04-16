FROM caddy:2.11.2-builder@sha256:92adff27ec7e96a6dc42324c3b89a5b1402bfe9ea95b214639ab3015c4f4f6f5 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.2@sha256:2acb10cebb92eea91a40b76691aff73adde9151416facbeab630bbc66d0969ab

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
