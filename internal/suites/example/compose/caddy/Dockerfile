FROM caddy:2.10.2-builder@sha256:13bf132b50ab5a3abd4d62d4da27a9d8fd98818028746fe85e745d73e3e71f3d AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:4163a5c7b7631707956db4057720ec75de429992d5e3aa518d54872c01644dbe

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
