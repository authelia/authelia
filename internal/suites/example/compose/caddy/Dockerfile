FROM caddy:2.9.1-builder@sha256:4acea926083a7bc7a4aca4250cec68c9dd6d4a6dc66c6ca8e26ceb1041bca86a AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:f27c4c64e20ca80651d1fdee831785e49553169bf9c9668600b456932f75ac47

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
