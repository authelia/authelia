FROM haproxy:3.3.0-alpine@sha256:bda550a1038ef57a62f8a03eba3dd389a50774738efaccceefbb7189b3286f3f
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
