FROM alpine:3.21.0@sha256:e323a465c03a31ad04374fc7239144d0fd4e2b92da6e3e0655580476d3a84621

RUN \
apk add --no-cache \
  bash \
  krb5 \
  openldap-clients \
  samba-dc \
  supervisor

CMD /init.sh setup
