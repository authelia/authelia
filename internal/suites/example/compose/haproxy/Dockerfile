FROM haproxy:3.4.0-alpine@sha256:a55a53ca3e7d9a8dbdd09a124b1e0390dacc14e663addec6e29c5a63f02e8647
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
