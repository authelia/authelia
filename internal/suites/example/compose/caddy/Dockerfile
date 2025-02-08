FROM caddy:2.9.1-builder@sha256:3c1a90ada80f1b9239c2248a3b3f846501849fc8d934e288442ae4456c82a1e1 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:a863d46cf06a9084f36cbffbe9f4ad046971dca32f79c68129aaf15ad356d6ce

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
