FROM caddy:2.8.4-builder@sha256:b1ee6157da9ce89796a6ce266713d701b383a17db2fa9cdb6dac04ef157544c8 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:1841e7c656154710f2fec273e12d4e517eeea5bd7a6c75e01dd26b88aaba9646

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
