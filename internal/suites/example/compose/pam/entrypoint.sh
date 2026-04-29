#!/bin/sh
set -e

AUTHELIA_URL="${AUTHELIA_URL:-https://login.example.com:8080}"
CA_CERT="${CA_CERT:-}"

CA_FLAG=""
if [ -n "${CA_CERT}" ]; then
    CA_FLAG=" ca-cert=${CA_CERT}"
fi

PAM_COMMON="account required pam_permit.so
session required pam_permit.so"

cat > /etc/pam.d/authelia-1fa <<EOF
auth required pam_authelia.so url=${AUTHELIA_URL} auth-level=1FA cookie-name=authelia_session${CA_FLAG} debug
${PAM_COMMON}
EOF

cat > /etc/pam.d/authelia-2fa <<EOF
auth required pam_unix.so
auth required pam_authelia.so url=${AUTHELIA_URL} auth-level=2FA cookie-name=authelia_session${CA_FLAG} debug
${PAM_COMMON}
EOF

cat > /etc/pam.d/authelia-1fa2fa <<EOF
auth required pam_authelia.so url=${AUTHELIA_URL} auth-level=1FA+2FA cookie-name=authelia_session${CA_FLAG} debug
${PAM_COMMON}
EOF

cat > /etc/pam.d/authelia-device-auth <<EOF
auth required pam_authelia.so url=${AUTHELIA_URL} auth-level=1FA+2FA cookie-name=authelia_session${CA_FLAG} method-priority=device_authorization oauth2-client-id=device-code oauth2-client-secret=foobar oauth2-scope=openid,authelia.pam timeout=3 debug
${PAM_COMMON}
EOF

cat > /etc/pam.d/authelia-device-auth-bind <<EOF
auth required pam_authelia.so url=${AUTHELIA_URL} auth-level=1FA+2FA cookie-name=authelia_session${CA_FLAG} method-priority=device_authorization oauth2-client-id=device-code oauth2-client-secret=foobar oauth2-scope=openid,authelia.pam timeout=120 debug
${PAM_COMMON}
EOF

# Default sshd PAM config to 1FA+2FA; tests cp the appropriate file to switch modes.
cp /etc/pam.d/authelia-1fa2fa /etc/pam.d/sshd

cat <<EOF
============================================
  Authelia PAM Suite Test Container
============================================
  Authelia URL:  ${AUTHELIA_URL}
  CA Cert:       ${CA_CERT:-<system default>}
  SSH Host:      ssh.example.com (192.168.240.130)
  SSH Port:      22
  SSH User:      john

  Default auth-level is 1FA+2FA. Tests switch modes by copying
  /etc/pam.d/authelia-{1fa,2fa,1fa2fa,device-auth,device-auth-bind} to /etc/pam.d/sshd.

  Connect with:
    ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\
        -o PreferredAuthentications=keyboard-interactive -o PubkeyAuthentication=no \\
        john@ssh.example.com
============================================
EOF

exec /usr/sbin/sshd.pam -D -e
