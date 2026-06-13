FROM alpine:3.24.0@sha256:a2d49ea686c2adfe3c992e47dc3b5e7fa6e6b5055609400dc2acaeb241c829f4

RUN \
echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
apk add --no-cache \
	bash \
	krb5 \
	openldap-clients \
	openssl \
	samba-dc \
	supervisor

CMD ["/init.sh", "setup"]
