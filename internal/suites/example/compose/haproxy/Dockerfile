FROM haproxy:3.3.1-alpine@sha256:48a5e38c6858fe479c727ebc49984ada607b73c76de2425486dcad4f3df02e1b
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
