FROM haproxy:3.3.4-alpine@sha256:f0419fc06ad76571a4ede85095e298d121cedd3a48add045d0b4a3c56df7366c
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
