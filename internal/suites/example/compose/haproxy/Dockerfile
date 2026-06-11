FROM haproxy:3.4.0-alpine@sha256:6306449418892136d0d9b768948972b11eebfe8b9f60594d261933642bddfb98
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
