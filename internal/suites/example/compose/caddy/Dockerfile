FROM caddy:2.11.2-builder@sha256:c54715e63a99f6bc427f6733749b71dd48fc036f9c4fcda10aa29fb278518ee0 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.2@sha256:1e40b251ca9639ead7b5cd2cedcc8765adfbabb99450fe23f130eefabf50f4bc

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
