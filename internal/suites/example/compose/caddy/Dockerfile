FROM caddy:2.10.0-builder@sha256:c7237a4e4b7d31b697a8672c7418be98471a3f30f44477e3fa36b9d16df35797 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:cd038960363e661b805a40ee6d03ba98608538f197c08cb915cb4af89e23a5b4

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
