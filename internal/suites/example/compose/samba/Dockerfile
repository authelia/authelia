# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM alpine:3.24.3@sha256:1b7232cda4ca3dc20b5d1247bdc9bfb631f2661d15d323bdd6fccc035ed1a9e7

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
