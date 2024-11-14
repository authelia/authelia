FROM caddy:2.8.4-builder@sha256:6e9733694e9d257c215925b2429c6e9949114ef7a4205d9e4718aefbf15daa77 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:cb50cf1d7c80d0dd310ce9ab0f2726706d82d3390bbd617bc213ef00377a8a70

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
