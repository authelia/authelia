FROM haproxy:3.4.3-alpine@sha256:fb87fc81943143b9acaea7442973e6ba654035fff76ffe7af6829dd1bcb0f7a5
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
