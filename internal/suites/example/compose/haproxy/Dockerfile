FROM haproxy:3.2.7-alpine@sha256:3b80483d47e1c7d1fc7eb4b9104f33d9a51259769be299eb675524dca2bc8157
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
