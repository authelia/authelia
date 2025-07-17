FROM haproxy:3.2.3-alpine@sha256:61d42cdaf13023e7f49e2a79b53b5a779e89f6ef341403fbbeaa91742cd0d20e
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
