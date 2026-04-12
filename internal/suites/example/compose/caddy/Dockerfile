FROM caddy:2.11.2-builder@sha256:5af59f1fa4f93a8945384cdd657ef879cac5034142bf2be9c489e2a570884673 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.2@sha256:1e40b251ca9639ead7b5cd2cedcc8765adfbabb99450fe23f130eefabf50f4bc

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
