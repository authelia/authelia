FROM haproxy:3.2.8-alpine@sha256:8c6b55d891584afded3ab1b91063a703ecc8ff2481302b691773648ffa806226
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
