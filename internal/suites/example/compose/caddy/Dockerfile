FROM caddy:2.10.2-builder@sha256:3c711fcf9a64d31e8927a2efb538cf99e6f758f737aadcde5fd7dfadc93287a8 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:4163a5c7b7631707956db4057720ec75de429992d5e3aa518d54872c01644dbe

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
