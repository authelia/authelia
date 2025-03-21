FROM haproxy:3.1.6-alpine@sha256:ad990d28b472c679d0045a2e84a3d98199d43e223fbab650b8cda1cbcc03bd1f
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
