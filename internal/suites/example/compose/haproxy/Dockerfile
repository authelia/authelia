FROM haproxy:3.1.3-alpine@sha256:2c8a8ee5503eb6712dd3f5a4c5bbea9de3d1c3a99f55ebdd61370c98de072c09
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
