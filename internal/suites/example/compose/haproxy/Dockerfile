FROM haproxy:3.2.7-alpine@sha256:b9c587547c9171dcc755c3b16188f24ed3cd412dc10c4e32050f642c9ee31c15
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
