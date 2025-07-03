FROM alpine:3.22.0@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715

RUN \
apk add --no-cache \
	bash \
	krb5 \
	openldap-clients \
	openssl \
	samba-dc \
	supervisor

CMD /init.sh setup
