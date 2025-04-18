FROM haproxy:3.1.7-alpine@sha256:3e1367158e93d65d0186d6b2fb94b0a5a5d7e1cac0cabedb0cda52c80dad1113
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
