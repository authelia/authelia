FROM haproxy:3.3.6-alpine@sha256:744be2dca649a44d490a4c565d36968d19482dd387f1bdd44c168f4322bc6b1e
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
