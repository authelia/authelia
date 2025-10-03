FROM haproxy:3.2.6-alpine@sha256:eb37f58646a901dc7727cf448cae36daaefaba79de33b5058dab79aa4c04aefb
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
