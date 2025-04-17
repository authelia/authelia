FROM haproxy:3.1.7-alpine@sha256:e0933d8fa6931244ecdd761c1be0fd02e11846551e5d0cb97229cfc2bf88e1ed
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
