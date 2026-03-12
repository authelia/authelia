FROM haproxy:3.3.5-alpine@sha256:debe69a8edcea8e8f6e240c346ba19a6fa0bd86156285a36aa888344a8604f5e
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
