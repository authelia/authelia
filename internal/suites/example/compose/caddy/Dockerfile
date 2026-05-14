FROM caddy:2.11.3-builder@sha256:f96a3b748f2ce4e5f6595453615da734b93993b231213fe35d0673893b5613ef AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.3@sha256:a464e5c4fb9c324b53adca428a40a66631e59127153b8816fab58103817ba8f0

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
