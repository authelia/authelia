# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM --platform=${BUILDPLATFORM} maven:3.10.0-eclipse-temurin-21@sha256:9b4877723dadf350b452dd97d9a6401e7b56f98fa9dfe420c32ad909989c7e4c AS builder

ARG CONFORMANCE_SUITE_REPO=https://gitlab.com/openid/conformance-suite.git
ARG CONFORMANCE_SUITE_REF=release-v5.3.2
ARG CONFORMANCE_SUITE_COMMIT=11999aad62ec292f36101410d9d5b761d44903ff

WORKDIR /src

COPY --link example/compose/oidc-conformance/authelia-login.patch /tmp/authelia-login.patch

RUN --mount=type=cache,target=/root/.m2 <<EOF
set -eu
git clone --depth 1 --branch "${CONFORMANCE_SUITE_REF}" "${CONFORMANCE_SUITE_REPO}" .
test "$(git rev-parse HEAD)" = "${CONFORMANCE_SUITE_COMMIT}"
git apply /tmp/authelia-login.patch
mvn -B package -DskipTests=true
EOF

FROM registry.gitlab.com/openid/conformance-suite:release-v5.3.2@sha256:c795df1b151e0c2f2e40f17a2470991cc38bc91628449e9bb090e2372fe8b822

COPY --link --from=builder /src/target/fapi-test-suite.jar /server/fapi-test-suite.jar

# The conformance server fetches Authelia's discovery document, JWKS, token endpoint and userinfo endpoint over TLS
# signed by the suite development CA, which no public root vouches for. Importing into the JDK's own cacerts adds that
# CA while leaving the public roots in place, which a -Djavax.net.ssl.trustStore override would not.
COPY --link common/pki/ca.public.crt /tmp/authelia-ca.crt

RUN keytool -importcert -noprompt -alias authelia-development-ca -file /tmp/authelia-ca.crt -cacerts -storepass changeit
