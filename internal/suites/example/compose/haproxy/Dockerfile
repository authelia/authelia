FROM haproxy:3.1.3-alpine@sha256:d34ef3e3ff220634bf59d2ca4e21cc6bae6addc09e20b3c1577b9bca351bbfc4
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
