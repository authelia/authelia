FROM caddy:2.8.4-builder@sha256:0c4ea007c72ccb87fad94e7cc404f02ae529eef98ceb488f42a6c87101905ab2 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:23203c92c239f04e7580b68af636db6bf1077e02dc41d23ee5c1dd259e4a562f

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
