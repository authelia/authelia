# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM registry.gitlab.com/openid/conformance-suite:release-v5.2.4@sha256:3a2615ed95a7f3bb92d545b4c65c0268f82a3893d6cd97fd98fd8e44eb15d81f

# The conformance server fetches Authelia's discovery document, JWKS, token endpoint and userinfo endpoint over TLS
# signed by the suite development CA, which no public root vouches for. Importing into the JDK's own cacerts adds that
# CA while leaving the public roots in place, which a -Djavax.net.ssl.trustStore override would not.
COPY --link common/pki/ca.public.crt /tmp/authelia-ca.crt

RUN keytool -importcert -noprompt -alias authelia-development-ca -file /tmp/authelia-ca.crt -cacerts -storepass changeit
