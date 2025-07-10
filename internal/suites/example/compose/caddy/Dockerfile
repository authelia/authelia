FROM caddy:2.10.0-builder@sha256:e5612b84a4a8c0d452da55d3d7d97985a5039ef3be6e399fd7264bb070f30559 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:f880657326733fb266a317b4ee31d6a3090a6bbff3dcb0c97833093810d59b46

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
