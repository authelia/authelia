FROM caddy:2.8.4-builder@sha256:03e3886f7b2af67ece93542d1e36d25ab0081166c4660411f661fac6df2bc1d6 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:cb50cf1d7c80d0dd310ce9ab0f2726706d82d3390bbd617bc213ef00377a8a70

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
