FROM caddy:2.11.2-builder@sha256:c94501daf39b91da9d8dcda6efffc40b6ae011a2530407872abea2df9ea4365c AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.2@sha256:25cdc846626b62d05f6b633b9b40c2c9f6ef89b515dc76133cefd920f7dbe562

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
