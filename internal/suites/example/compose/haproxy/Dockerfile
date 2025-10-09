FROM haproxy:3.2.6-alpine@sha256:3a819392f5a6af2a9203d1c63ff0feba74f0c940623a558511be06d7d85e361c
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
