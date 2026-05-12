FROM haproxy:3.3.10-alpine@sha256:3ea6cce8d7db321a9382445d2d13b2511ea5809036fdb4dd929ad2ec68cdfc84
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
