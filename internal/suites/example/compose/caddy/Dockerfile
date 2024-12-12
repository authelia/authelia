FROM caddy:2.8.4-builder@sha256:dd54498b9546c06cdafd0e3b20276e6999f07102cc1fb578f0408c3bf8399393 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:d17c155b627f4ae14cef9cb4143b63c529a8497966b62febcde79f4ecc3857f7

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
