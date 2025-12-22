FROM haproxy:3.3.1-alpine@sha256:cbfee3170c6fefb31b69b6eb3cf12ae5fc0747c838c70feb207aca651f30eba5
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
