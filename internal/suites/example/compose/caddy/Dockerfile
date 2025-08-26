FROM caddy:2.10.2-builder@sha256:22ff419120e17d257b9744216ecae0b2305d81ff570490fba03625f047c4195a AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:7f2af82ffc1849dff5f4b9d9b0405eae77ed6cebe6740b181ae72a668e3ed525

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
