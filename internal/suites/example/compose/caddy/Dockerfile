FROM caddy:2.9.1-builder@sha256:57f45106467c6368822bf1bd20b335b3c41102212d350e813a691f8978151814 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:a863d46cf06a9084f36cbffbe9f4ad046971dca32f79c68129aaf15ad356d6ce

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
