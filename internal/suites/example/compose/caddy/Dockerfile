FROM caddy:2.9.1-builder@sha256:4199ffa369bfa56693d6dce853dbf294c0472ddceb629127b8085e6f0457a281 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:116bea95dd70e611b4b46f5fc0fc931aa1b2acc75c0e8903c88a93a713d57e6d

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
