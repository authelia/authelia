FROM caddy:2.9.1-builder@sha256:84fcb47a54c6149dcbb8f6320dc9421a8c9fb16879d3b6c71741d6a9ebea7c37 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:cd261fc62394f1ff0b44f16eb1d202b4e71d5365c9ec866a4f1a9c5a52da9352

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
