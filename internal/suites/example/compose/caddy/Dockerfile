FROM caddy:2.10.2-builder@sha256:6b9203b6c8916b617cd00a43f9687ced89c3c8dac46cc2149cdab1c8703575ee AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:cef7fe17d4df3b0843eae25bed319d6140b1a7bb13ccef076f5f1783d5bca9b1

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
