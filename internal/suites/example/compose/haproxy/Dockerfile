FROM haproxy:3.3.1-alpine@sha256:609201fa88d66b3686b77510d7fa316c03f7498699f4523009ba2687db2423e9
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
