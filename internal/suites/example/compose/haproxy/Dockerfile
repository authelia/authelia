FROM haproxy:3.3.6-alpine@sha256:76f24e0d7723f45d260c9606c47f21efa1088ec384d12a89c3ff3ec93165e577
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
