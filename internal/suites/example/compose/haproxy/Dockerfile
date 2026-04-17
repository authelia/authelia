FROM haproxy:3.3.6-alpine@sha256:4f97a2cb7f02fd08402259e74a65ef12fcfa3dff1ef78fddecb5228a17b7f4ad
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
