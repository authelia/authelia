FROM caddy:2.10.2-builder@sha256:a642fca0473850425b8b62c9ab33fe14306ec80eb27a007f5f893e7357569451 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:70e816c44fb79071fc4cd939ffda76e3b629642309efe31a4fb0ed45873be464

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
