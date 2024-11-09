FROM haproxy:3.0.6-alpine@sha256:fe6f73e6c40d762d44b36dc77a0b650bba138022803bf6c8f94d9f17f186d958
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
