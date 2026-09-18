# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM alpine:3.24.2@sha256:31b6477333eb8257db9e5d7c3a7264fd0467928756f0bbcc27d35bea5d28cdbd

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
