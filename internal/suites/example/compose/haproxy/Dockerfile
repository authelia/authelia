FROM haproxy:3.3.1-alpine@sha256:d3b50f8bc0121e49b47ce4118571da140f8a62b38f6ca011b66ebfd09750c87a
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
