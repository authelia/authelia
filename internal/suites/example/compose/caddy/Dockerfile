FROM caddy:2.9.1-builder@sha256:3f2edfec10ca2baa71333220af869c0c2f6ccdce37ff502542152c2d8f00e05d AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:a863d46cf06a9084f36cbffbe9f4ad046971dca32f79c68129aaf15ad356d6ce

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
