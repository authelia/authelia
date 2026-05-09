FROM haproxy:3.3.9-alpine@sha256:0b99bdae555b9483d98250fd8c0d1a8bdf897d17894ea0cdfbccbab89ef3c4bd
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
