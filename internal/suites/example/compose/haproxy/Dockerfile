FROM haproxy:3.2.4-alpine@sha256:436a5691e6fd563ca4090e351048fda1066bec93b595cc8eaf26988c65eb1152
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
