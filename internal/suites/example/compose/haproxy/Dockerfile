FROM haproxy:3.0.4-alpine@sha256:e4d011524b6a9a530c97114d090a3fbac1c6a70e24beb02c7ca740efd8ebdc5f
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
