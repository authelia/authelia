FROM haproxy:3.1.3-alpine@sha256:0a4ad07f0149bf39da70348464592769e99a6dea80df86cc95b725268a1e5585
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
