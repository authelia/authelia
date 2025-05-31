FROM haproxy:3.2.0-alpine@sha256:a0345cd0591745815d30f7674b362d018e1c04c5314beb3bd87f02387c9dc236
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
