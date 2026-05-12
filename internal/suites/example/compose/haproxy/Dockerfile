FROM haproxy:3.3.10-alpine@sha256:3201152205e9d02557ab3c8ecfcf0d378126eec0a5d1b0950ea22af31ee10bf2
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
