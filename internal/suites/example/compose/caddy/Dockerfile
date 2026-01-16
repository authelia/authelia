FROM caddy:2.10.2-builder@sha256:f556cebc3c90b8fdc068364e13c320d460ed9b3445e46dda6589c7993a7d115c AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:86d2ab93167b0c24f9f86e8a4214dee0ba25f86808002bfd29a78677824a2483

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
