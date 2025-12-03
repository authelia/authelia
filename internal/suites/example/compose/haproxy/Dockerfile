FROM haproxy:3.3.0-alpine@sha256:3755973a59dd758b24d952649ca1f003d744b9e771f61f5e7db63e47baa68e3b
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
