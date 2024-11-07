FROM caddy:2.8.4-builder@sha256:e8ae2ab1723d2c4d8479950bd5af1955371c7a92bb097c0d40e1550f090b28b3 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:63d8776389cc6527e4a23bd9750489dc661923cffc3b9d7e0c20e062fa0325ec

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
