FROM haproxy:3.0.3-alpine@sha256:4b2c1d41d0448b9baf60186995279f71fb36552a31ff028434224d73f352b65e
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
