FROM haproxy:3.2.9-alpine@sha256:a92591a27182bb3207fc5770c5480107f13b2978ac5599248c3f228057f7e7d1
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
