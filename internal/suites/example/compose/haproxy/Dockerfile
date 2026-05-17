FROM haproxy:3.3.10-alpine@sha256:0b1bd52a831d0f0e0fb03a56e59fcd5c16b8e219b0e5f2eff388b40fd95b5d44
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
