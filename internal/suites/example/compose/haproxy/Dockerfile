FROM haproxy:3.1.1-alpine@sha256:e560774ab8608e38ae03ac52062e923dcb3c71c5bd2f55ad22a1521d7e57e725
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
