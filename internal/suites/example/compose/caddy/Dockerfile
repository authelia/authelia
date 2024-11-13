FROM caddy:2.8.4-builder@sha256:0c4ea007c72ccb87fad94e7cc404f02ae529eef98ceb488f42a6c87101905ab2 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:827d2118f9164f10df89c80c81e89666f031459945cce78b2c4917a8a8a0a0ab

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
