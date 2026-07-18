FROM haproxy:3.4.2-alpine@sha256:0878b11eb64c433be1b0f578a584b8aca12f6caaa64c8f239b8b556c0dd5eeeb
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
