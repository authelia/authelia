FROM haproxy:3.2.0-alpine@sha256:2de6e08f57cd518c2f8e6d691e297477eff1113f7f6cc9ec314ea0d02b72195d
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
