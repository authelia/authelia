FROM haproxy:3.4.4-alpine@sha256:e37e968e65707ae577e3aaa09ee181b605dd8562beabe982dcd837f9fd5fdf83
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
