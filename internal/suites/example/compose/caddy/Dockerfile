FROM caddy:2.10.2-builder@sha256:fe404674d209455fdef351db5437758ee0e70a6b59abe770663c09cfa05dbddf AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:dedfbbeb703b2ce9ff4a98fc06aa9c7c7d9a42f0b7d778738c1dd3ef11dcc767

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
