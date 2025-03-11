FROM haproxy:3.1.5-alpine@sha256:5090db844ab40e648ad485eefc7cf469b4aee3c9fe6c4c65cbb19d99d13a0952
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
