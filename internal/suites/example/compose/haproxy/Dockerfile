FROM haproxy:3.1.1-alpine@sha256:2adc8518931611c911633d85881776566c6d5e03658a6f13fa772abdf616ae33
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
