FROM caddy:2.9.1-builder@sha256:3f2edfec10ca2baa71333220af869c0c2f6ccdce37ff502542152c2d8f00e05d AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:116bea95dd70e611b4b46f5fc0fc931aa1b2acc75c0e8903c88a93a713d57e6d

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
