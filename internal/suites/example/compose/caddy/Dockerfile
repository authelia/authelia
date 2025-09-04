FROM caddy:2.10.2-builder@sha256:2c02c4fadac9cd4e100155596d89913ecaa9440ddc6507f16c7bdb6489f8d9a7 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:4163a5c7b7631707956db4057720ec75de429992d5e3aa518d54872c01644dbe

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
