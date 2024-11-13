FROM caddy:2.8.4-builder@sha256:03e3886f7b2af67ece93542d1e36d25ab0081166c4660411f661fac6df2bc1d6 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:827d2118f9164f10df89c80c81e89666f031459945cce78b2c4917a8a8a0a0ab

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
