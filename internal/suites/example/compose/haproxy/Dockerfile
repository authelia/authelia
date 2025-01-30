FROM haproxy:3.1.2-alpine@sha256:1dbf50f791f09f0c392a18b8278baeee9a4bedeefa13b7ccf7c95e5cca95ea2f
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
