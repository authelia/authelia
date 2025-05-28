FROM haproxy:3.2.0-alpine@sha256:93ad638f0dc5ad6c0c2a8115cef152cde43cb1f9f43dd881224d41fea3dd17c1
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
