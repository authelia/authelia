# Builds the pam_authelia Go binary and the pam_authelia.so C shim by cloning
# https://github.com/authelia/pam at the specified ref. Override at build time
# to pin a branch, tag, or commit for local testing:
#   docker build --build-arg AUTHELIA_PAM_REF=feat-foo \
#                --build-arg AUTHELIA_PAM_REPO=https://github.com/authelia/pam.git

FROM golang:1.26.2-alpine AS builder

RUN <<EOF
set -eu
apk --no-cache add git gcc musl-dev linux-pam-dev make
EOF

ARG AUTHELIA_PAM_REPO=https://github.com/authelia/pam.git
ARG AUTHELIA_PAM_REF=feat-pam

RUN <<EOF
set -eu
git clone ${AUTHELIA_PAM_REPO} /src
cd /src
git checkout ${AUTHELIA_PAM_REF}
EOF

WORKDIR /src
RUN <<EOF
set -eu
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /pam_authelia ./cmd/pam_authelia/
EOF

RUN <<EOF
set -eu
make -C shim
EOF

FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659

RUN <<EOF
set -eu
apk --no-cache add linux-pam openssh-server-pam ca-certificates
mkdir -p /var/run/sshd
ssh-keygen -A
adduser -D -s /bin/sh john
echo "john:password" | chpasswd
sed -i 's/^#*KbdInteractiveAuthentication.*/KbdInteractiveAuthentication yes/' /etc/ssh/sshd_config
sed -i 's/^#*UsePAM.*/UsePAM yes/' /etc/ssh/sshd_config
sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config
echo "AuthenticationMethods keyboard-interactive" >> /etc/ssh/sshd_config
EOF

COPY --link --from=builder /pam_authelia /usr/bin/pam_authelia
COPY --link --from=builder /src/shim/pam_authelia.so /tmp/pam_authelia.so

RUN <<EOF
set -eu
PAM_DIR="$(dirname "$(find /lib /usr/lib -name 'pam_unix.so' -print -quit 2>/dev/null)")"
cp /tmp/pam_authelia.so "${PAM_DIR}/pam_authelia.so"
rm /tmp/pam_authelia.so
EOF

COPY --link --chmod=755 internal/suites/example/compose/pam/entrypoint.sh /entrypoint.sh

EXPOSE 22

ENTRYPOINT ["/entrypoint.sh"]
