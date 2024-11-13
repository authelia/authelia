FROM caddy:2.8.4-builder@sha256:a55a78f46e30bd6cc090fd9936669d7793b05e356f1382d13df5b3960ed99d31 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:827d2118f9164f10df89c80c81e89666f031459945cce78b2c4917a8a8a0a0ab

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
