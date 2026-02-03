FROM caddy:2.10.2-builder@sha256:23e70f63a88bacd9b0564d055eed5c25b98e90930876cca01aee1f5d1ae29748 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:70e816c44fb79071fc4cd939ffda76e3b629642309efe31a4fb0ed45873be464

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
