FROM haproxy:3.4.2-alpine@sha256:36641a99dc71dfdeea2b99e0027dfe5a8c1ba3eb3f6f9049479cacfd4720558d
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
