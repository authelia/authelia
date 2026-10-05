# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM --platform=${BUILDPLATFORM} maven:3.10.0-eclipse-temurin-21@sha256:fe1dc38166b19b780a9f6195576da54eab69a9f9ef1e706e617e67197ea0d1e1 AS builder

ARG CONFORMANCE_SUITE_REPO=https://gitlab.com/openid/conformance-suite.git
ARG CONFORMANCE_SUITE_REF=release-v5.3.1
ARG CONFORMANCE_SUITE_COMMIT=440eec8bac7b12b7389d7ca9cbc459b53507a443

WORKDIR /src

COPY --link example/compose/oidc-conformance/authelia-login.patch /tmp/authelia-login.patch

RUN --mount=type=cache,target=/root/.m2 <<EOF
set -eu
git clone --depth 1 --branch "${CONFORMANCE_SUITE_REF}" "${CONFORMANCE_SUITE_REPO}" .
test "$(git rev-parse HEAD)" = "${CONFORMANCE_SUITE_COMMIT}"
git apply /tmp/authelia-login.patch
mvn -B package -DskipTests=true
EOF

FROM registry.gitlab.com/openid/conformance-suite:release-v5.3.1@sha256:69495f453a920c262f66e5e72abd12501c33e05ce88051cddf300c00621a4d70

COPY --link --from=builder /src/target/fapi-test-suite.jar /server/fapi-test-suite.jar

# The conformance server fetches Authelia's discovery document, JWKS, token endpoint and userinfo endpoint over TLS
# signed by the suite development CA, which no public root vouches for. Importing into the JDK's own cacerts adds that
# CA while leaving the public roots in place, which a -Djavax.net.ssl.trustStore override would not.
COPY --link common/pki/ca.public.crt /tmp/authelia-ca.crt

RUN keytool -importcert -noprompt -alias authelia-development-ca -file /tmp/authelia-ca.crt -cacerts -storepass changeit
