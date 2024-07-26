FROM caddy:2.8.4-builder@sha256:efb0a066246f690811b6063d74e5ba633f62286ed15f39f6eb3d05c2a870a2fe AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:b8c90e126558630ac99b6df3daf4aa261e80dd0bd28f8b4882be4f31d783706d

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
