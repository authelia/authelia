FROM haproxy:3.3.7-alpine@sha256:2afa53c856e4e9fcc7dfb35b807fcb189896d7e62b38d363f9bedea92bce7f9a
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
