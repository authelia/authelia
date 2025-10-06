FROM haproxy:3.2.6-alpine@sha256:923d441e105ed876cf6f6bb56a5a1b78e95bc89f3bdbf3611c18f26bc4e3f7f2
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
