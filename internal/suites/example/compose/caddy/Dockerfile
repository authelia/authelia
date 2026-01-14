FROM caddy:2.10.2-builder@sha256:efb93f7af011ef86c2eea10700a1a6fa8d1d99228f2dd94f7c2a12192e44d3c9 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:86d2ab93167b0c24f9f86e8a4214dee0ba25f86808002bfd29a78677824a2483

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
