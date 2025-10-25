FROM haproxy:3.2.7-alpine@sha256:692016a80e568fe39546267518909d236e6bf88f98996a1c03c98156e7f6f315
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
