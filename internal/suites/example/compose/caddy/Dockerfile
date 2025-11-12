FROM caddy:2.10.2-builder@sha256:7838e973e96b0e5f74213bf9f3c0f35089dc2e812e4c0b20790e45c6512423fb AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:cef7fe17d4df3b0843eae25bed319d6140b1a7bb13ccef076f5f1783d5bca9b1

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
