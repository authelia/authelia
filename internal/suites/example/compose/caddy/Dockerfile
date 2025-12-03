FROM caddy:2.10.2-builder@sha256:fe404674d209455fdef351db5437758ee0e70a6b59abe770663c09cfa05dbddf AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:cef7fe17d4df3b0843eae25bed319d6140b1a7bb13ccef076f5f1783d5bca9b1

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
