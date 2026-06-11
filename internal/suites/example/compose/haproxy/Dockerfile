FROM haproxy:3.4.0-alpine@sha256:d0ca281fca326f9283198afc25c53525cfc6b44190b441882b279241ca57d5d0
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
