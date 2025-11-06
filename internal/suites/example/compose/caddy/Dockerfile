FROM caddy:2.10.2-builder@sha256:4535684b970f048a772efec86ce46148a0e916a653e73d9b223cb2ad07f5dfe6 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:9e95012adcbbd599f853cb315b986781845c238f9e52aa3652798758cca01422

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
