FROM caddy:2.10.0-builder@sha256:cc39c328df6f5688c9fd98c2f4babb9b6ba38f1dda7d9a47b726fbc5b863d655 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:c10cf853496a180595e521a7e5081e0ba0a85e4bd54321a751e3dc91c7b3dac9

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
