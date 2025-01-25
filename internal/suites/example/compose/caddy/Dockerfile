FROM caddy:2.8.4-builder@sha256:9f86856dd089cd85f546b6a022079c9f157ab62bc1d688ecd9b5d474ea81a487 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:226d1f059b75399fe19182893c7184591c07b97afc8dfcf44eeb80c9a77a530f

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
