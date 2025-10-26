FROM haproxy:3.2.7-alpine@sha256:d7fa3863dbf083c4c2e421ab55755c9af67ac09542233660bc3342a8f70ac567
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
