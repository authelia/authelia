FROM caddy:2.10.2-builder@sha256:fbcbf6ec1f6d94bc85be547a4d46ac32d0d8119baeaffc383dee068ba332cb50 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:c3d7ee5d2b11f9dc54f947f68a734c84e9c9666c92c88a7f30b9cba5da182adb

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
