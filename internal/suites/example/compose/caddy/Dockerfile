FROM caddy:2.9.1-builder@sha256:4199ffa369bfa56693d6dce853dbf294c0472ddceb629127b8085e6f0457a281 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:f27c4c64e20ca80651d1fdee831785e49553169bf9c9668600b456932f75ac47

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
