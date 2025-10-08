FROM caddy:2.10.2-builder@sha256:4d2be51cdd15fc6872738bc1f26f7c0606c8bd135e7580d4fe6cf54810b19bb7 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:87aa104ed6c658991e1b0672be271206b7cd9fec452d1bf3ed9ad6f8ab7a2348

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
