FROM haproxy:3.0.3-alpine@sha256:b6f165eccb512048a66a63b67edb2a5f613a97b6227c317a92b3905c74bde6e5
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
