FROM haproxy:3.4.0-alpine@sha256:4bb437019ee4bbf457c583d8dcdac6469ea750f1d8d7f013858a2a5ad92bd10d
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
