FROM haproxy:3.2.4-alpine@sha256:acf1c21a011a985ccbfbf2f524e79fe5b656dea30ee7353221a38071749c0fe9
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
