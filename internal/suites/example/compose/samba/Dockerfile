FROM alpine:3.23.2@sha256:865b95f46d98cf867a156fe4a135ad3fe50d2056aa3f25ed31662dff6da4eb62

RUN \
echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
apk add --no-cache \
	bash \
	krb5@edge \
	openldap-clients \
	openssl \
	samba-dc@edge \
	supervisor

CMD ["/init.sh", "setup"]
