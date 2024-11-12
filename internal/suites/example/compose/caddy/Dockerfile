FROM caddy:2.8.4-builder@sha256:0c4ea007c72ccb87fad94e7cc404f02ae529eef98ceb488f42a6c87101905ab2 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:b5154b3538b1eb8f6df79fd9c15614e6ce0efd0ea7c6342047f061db95ac5cda

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
