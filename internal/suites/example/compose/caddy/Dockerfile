FROM caddy:2.10.2-builder@sha256:00dfa5f39333243282c82d73de2acfe0609cb4a7879eb96946dcdfbf3a27ff87 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:70e816c44fb79071fc4cd939ffda76e3b629642309efe31a4fb0ed45873be464

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
