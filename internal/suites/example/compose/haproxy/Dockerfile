FROM haproxy:3.1.1-alpine@sha256:eaf066f6b8a6b23286ee7d29ab827d8263023ed394f89f1d76ce06b2b031e9a0
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
