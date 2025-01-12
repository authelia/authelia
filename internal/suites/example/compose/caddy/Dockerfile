FROM caddy:2.8.4-builder@sha256:9f86856dd089cd85f546b6a022079c9f157ab62bc1d688ecd9b5d474ea81a487 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:d17c155b627f4ae14cef9cb4143b63c529a8497966b62febcde79f4ecc3857f7

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
