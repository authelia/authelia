FROM caddy:2.10.2-builder@sha256:4d2be51cdd15fc6872738bc1f26f7c0606c8bd135e7580d4fe6cf54810b19bb7 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:614bbc6da7ec42f3c76077e86f429297047680f9cb420ad0f07a8fe049193d89

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
