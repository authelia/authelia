FROM haproxy:3.3.7-alpine@sha256:e32228771ff0efed99814fbeb5e5de73fd6a81dedb90afb871c8cc4b57ac52bb
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
