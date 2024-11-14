FROM haproxy:3.0.6-alpine@sha256:c22c8710886104a48b920306f063401f0d11811858e3c6b9d87d88a7556b2e61
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
