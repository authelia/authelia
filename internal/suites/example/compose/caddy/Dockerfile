FROM caddy:2.10.2-builder@sha256:53f91ad7c5f1ab9a607953199b7c1e10920c570ae002aef913d68ed7464fb19f AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:9e95012adcbbd599f853cb315b986781845c238f9e52aa3652798758cca01422

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
