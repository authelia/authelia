FROM haproxy:3.2.3-alpine@sha256:9035998bb313087edf88b42c9868b41521e96ceeb122f15b345d9c9bcd716cb8
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
