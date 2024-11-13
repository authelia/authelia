FROM caddy:2.8.4-builder@sha256:0c4ea007c72ccb87fad94e7cc404f02ae529eef98ceb488f42a6c87101905ab2 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:6f3eea53e585bd8036028ca703de1c805353b79b43c8b3dd7168ac65827e7ef1

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
