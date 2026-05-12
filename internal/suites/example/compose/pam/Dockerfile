# Builds the pam_authelia Go binary and the pam_authelia.so C shim by cloning
# https://github.com/authelia/pam at the specified ref. Override at build time
# to pin a branch, tag, or commit for local testing:
#   docker build --build-arg AUTHELIA_PAM_REF=feat-foo \
#                --build-arg AUTHELIA_PAM_REPO=https://github.com/authelia/pam.git

FROM golang:1.26.3-alpine@sha256:91eda9776261207ea25fd06b5b7fed8d397dd2c0a283e77f2ab6e91bfa71079d AS builder

RUN <<EOF
set -eu
apk --no-cache add git gcc musl-dev linux-pam-dev make
EOF

ARG AUTHELIA_PAM_REPO=https://github.com/authelia/pam.git
ARG AUTHELIA_PAM_REF=master

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

FROM alpine:3.23.4@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11

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
