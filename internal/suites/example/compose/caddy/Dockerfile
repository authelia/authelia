FROM caddy:2.10.2-builder@sha256:b33f07bbf5f8a431f224b3df0f0e709f4c7cb484515df2cdf71218e6e6021205 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:4163a5c7b7631707956db4057720ec75de429992d5e3aa518d54872c01644dbe

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
