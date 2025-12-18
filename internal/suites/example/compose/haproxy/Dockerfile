FROM haproxy:3.3.0-alpine@sha256:42bda80555d496d2f7bf6a4cbca3fed2ec12bbda49bdfaef34c662bdfa6daff1
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
