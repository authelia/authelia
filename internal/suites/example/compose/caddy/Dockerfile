FROM caddy:2.8.4-builder@sha256:dd54498b9546c06cdafd0e3b20276e6999f07102cc1fb578f0408c3bf8399393 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:226d1f059b75399fe19182893c7184591c07b97afc8dfcf44eeb80c9a77a530f

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
