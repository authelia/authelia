FROM haproxy:3.2.2-alpine@sha256:86a23059cd1422671f6db178dfc37ea247505c6face3b55a60294b60048ecbd3
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
