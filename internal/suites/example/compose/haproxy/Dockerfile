FROM haproxy:3.2.1-alpine@sha256:ac79fe145f2bb6626ff26b584a2d0a34e791906c01015f2ae037aa3137b683d9
USER root

RUN \
apk add --no-cache \
  lua-json4

USER haproxy
