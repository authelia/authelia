FROM haproxy:3.4.1-alpine@sha256:820658d9303227313f1feb164443f07319d604fed68982c918611dcaea8fb775
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
