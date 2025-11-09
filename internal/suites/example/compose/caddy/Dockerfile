FROM caddy:2.10.2-builder@sha256:7838e973e96b0e5f74213bf9f3c0f35089dc2e812e4c0b20790e45c6512423fb AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:9e95012adcbbd599f853cb315b986781845c238f9e52aa3652798758cca01422

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
