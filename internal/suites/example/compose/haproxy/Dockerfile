FROM haproxy:3.3.2-alpine@sha256:86245f821fda607572cacf0b090baf4d5e9f3fe84019f240cfdf35f677b764cc
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
