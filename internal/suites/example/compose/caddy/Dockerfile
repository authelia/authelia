FROM caddy:2.9.1-builder@sha256:c1e258b449a2deaab94f90330450fcfe8c3e0b6f0891f28a704c749f43825877 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:cd261fc62394f1ff0b44f16eb1d202b4e71d5365c9ec866a4f1a9c5a52da9352

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
