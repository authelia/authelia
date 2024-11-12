FROM caddy:2.8.4-builder@sha256:0b67771fda573b19475cd37797852879f0446d5b9f9d93f0b1c81f80d659d71a AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:b5154b3538b1eb8f6df79fd9c15614e6ce0efd0ea7c6342047f061db95ac5cda

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
