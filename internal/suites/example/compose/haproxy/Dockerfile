FROM haproxy:3.4.0-alpine@sha256:f5437564f819793be4614085ee0bafae45d0579fa2b97f9c76449e60436476e0
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
