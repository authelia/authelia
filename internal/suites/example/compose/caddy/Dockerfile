FROM caddy:2.8.4-builder@sha256:202faaa291d9da41346fbb03609827e9e8004e7c74b3c1bbb5c1248575958da1 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:1841e7c656154710f2fec273e12d4e517eeea5bd7a6c75e01dd26b88aaba9646

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
