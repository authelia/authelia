FROM haproxy:3.3.8-alpine@sha256:10690acb357180d5214c6fce59e2cefded6cc72b0f7e3febb323fea95b27e349
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
