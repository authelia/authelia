FROM caddy:2.10.2-builder@sha256:6e3ed727ce8695fc58e0a8de8e5d11888f6463c430ea5b40e0b5f679ab734868 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:cef7fe17d4df3b0843eae25bed319d6140b1a7bb13ccef076f5f1783d5bca9b1

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
