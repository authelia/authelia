FROM haproxy:3.2.9-alpine@sha256:d32199ab40a0aa673a8a5d822c8c44e9a3d13def5c1d81dba0b9fbb7f2d4834e
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
