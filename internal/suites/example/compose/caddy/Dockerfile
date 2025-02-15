FROM caddy:2.9.1-builder@sha256:4199ffa369bfa56693d6dce853dbf294c0472ddceb629127b8085e6f0457a281 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:ca29b067a632213b310e82a2881a17f186b87913044eb6ed9ea632aeb0d36cea

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
