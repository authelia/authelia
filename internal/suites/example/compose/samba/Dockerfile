FROM alpine:3.17.3

RUN \
apk add --no-cache \
  bash \
  krb5 \
  openldap-clients \
  samba-dc \
  supervisor

CMD /init.sh setup
