FROM caddy:2.8.4-builder@sha256:fe45985056ead0e8226825284ccab1855ccdd4e9fc9c656175a494fe02b21c1a AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:63d8776389cc6527e4a23bd9750489dc661923cffc3b9d7e0c20e062fa0325ec

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
