FROM caddy:2.9.1-builder@sha256:4c455f2cf685637594f7112b2526229a58a6039975fd1915281d452b2075bda3 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:cd261fc62394f1ff0b44f16eb1d202b4e71d5365c9ec866a4f1a9c5a52da9352

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
