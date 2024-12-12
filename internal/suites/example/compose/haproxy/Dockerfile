FROM haproxy:3.1.0-alpine@sha256:19702fca249ef4ccceb0d970fb5ff2948cee47b707b180ae1ac25bf94c900f2c
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
