FROM haproxy:3.3.5-alpine@sha256:cb1b8fd0499174728d588af8cde5e8c366b4e1b99e31c4874a97c4a5fdeafa11
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
