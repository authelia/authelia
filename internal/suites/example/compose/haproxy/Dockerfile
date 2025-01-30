FROM haproxy:3.1.3-alpine@sha256:1b9a695f1e50d69fbc9d2cc5b124210a8124cfc9f354deedab2ff16955a9d3e4
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
