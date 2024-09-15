FROM haproxy:3.0.4-alpine@sha256:0a4fc48c3dda253dd6173fd412eb31048578f99e28e401da58375c3f0876bae1
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
