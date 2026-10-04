# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM --platform=${BUILDPLATFORM} maven:3.9.16-eclipse-temurin-21@sha256:99e61abcff91a9b1333463bd8451fb18495d6eba9250ac66a338b518f8278320 AS builder

ARG CONFORMANCE_SUITE_REPO=https://gitlab.com/openid/conformance-suite.git
ARG CONFORMANCE_SUITE_REF=release-v5.2.4
ARG CONFORMANCE_SUITE_COMMIT=ab35a8df4864da35b49eff11483e204e01aa7961

WORKDIR /src

COPY --link example/compose/oidc-conformance/authelia-login.patch /tmp/authelia-login.patch

RUN --mount=type=cache,target=/root/.m2 <<EOF
set -eu
git clone --depth 1 --branch "${CONFORMANCE_SUITE_REF}" "${CONFORMANCE_SUITE_REPO}" .
test "$(git rev-parse HEAD)" = "${CONFORMANCE_SUITE_COMMIT}"
git apply /tmp/authelia-login.patch
mvn -B package -DskipTests=true
EOF

FROM registry.gitlab.com/openid/conformance-suite:release-v5.2.4@sha256:3a2615ed95a7f3bb92d545b4c65c0268f82a3893d6cd97fd98fd8e44eb15d81f

COPY --link --from=builder /src/target/fapi-test-suite.jar /server/fapi-test-suite.jar

# The conformance server fetches Authelia's discovery document, JWKS, token endpoint and userinfo endpoint over TLS
# signed by the suite development CA, which no public root vouches for. Importing into the JDK's own cacerts adds that
# CA while leaving the public roots in place, which a -Djavax.net.ssl.trustStore override would not.
COPY --link common/pki/ca.public.crt /tmp/authelia-ca.crt

RUN keytool -importcert -noprompt -alias authelia-development-ca -file /tmp/authelia-ca.crt -cacerts -storepass changeit
