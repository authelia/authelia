FROM haproxy:3.0.5-alpine@sha256:43a2591a6a62a41dd22edb4452880e9c70de489fbd7c54ffb52d035cccbfe949
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
