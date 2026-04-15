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
auth required pam_authelia.so url=${AUTHELIA_URL} auth-level=1FA+2FA cookie-name=authelia_session${CA_FLAG} method-priority=device_authorization oauth2-client-id=device-code oauth2-client-secret=foobar timeout=3 debug
${PAM_COMMON}
EOF

# Default sshd PAM config to 1FA+2FA; tests cp the appropriate file to switch modes.
cp /etc/pam.d/authelia-1fa2fa /etc/pam.d/sshd

echo "============================================"
echo "  Authelia PAM Suite Test Container"
echo "============================================"
echo "  Authelia URL:  ${AUTHELIA_URL}"
echo "  CA Cert:       ${CA_CERT:-<system default>}"
echo "  SSH Host:      ssh.example.com (192.168.240.130)"
echo "  SSH Port:      22"
echo "  SSH User:      john"
echo ""
echo "  Default auth-level is 1FA+2FA. Tests switch modes by copying"
echo "  /etc/pam.d/authelia-{1fa,2fa,1fa2fa} to /etc/pam.d/sshd."
echo ""
echo "  Connect with:"
echo "    ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \\"
echo "        -o PreferredAuthentications=keyboard-interactive -o PubkeyAuthentication=no \\"
echo "        john@ssh.example.com"
echo "============================================"

exec /usr/sbin/sshd.pam -D -e
