FROM caddy:2.5.0-builder AS builder

RUN xcaddy build aae6ce27016cb3f2e5cd1d8b594b36c03edeeceb

FROM caddy:2.5.0

COPY --from=builder /usr/bin/caddy /usr/bin/caddy