FROM haproxy:3.2.5-alpine@sha256:8007effce89a08af0236b9529a0daab5b8b36fa939f4162f28201f1bf8731dbf
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
