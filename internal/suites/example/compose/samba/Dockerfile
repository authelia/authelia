FROM alpine:3.23.0@sha256:51183f2cfa6320055da30872f211093f9ff1d3cf06f39a0bdb212314c5dc7375

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
