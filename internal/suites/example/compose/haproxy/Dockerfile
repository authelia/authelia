FROM haproxy:3.1.3-alpine@sha256:54a9bda49fb275fb648afc5e4f81ed86218337434c22946c15247a681af033a8
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
