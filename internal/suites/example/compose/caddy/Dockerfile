FROM caddy:2.8.4-builder@sha256:8778619a8bfd796c5bafa16a1273e092612c57b34e25cb62d12ea6636f1ae252 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:63d8776389cc6527e4a23bd9750489dc661923cffc3b9d7e0c20e062fa0325ec

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
